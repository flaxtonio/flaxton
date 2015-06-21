/**
 * Created by tigran on 5/24/15.
 */


var mongoose = require("mongoose")
    , md5 = require("MD5")
    , User = mongoose.model("User")
    , Task = mongoose.model("Task")
    , Daemon = mongoose.model("Daemon")
    , Container = mongoose.model("Container")
    , StateLogger = mongoose.model("StateLogger")
    , async = require("async");

module.exports = {
    get_daemon: function(req, res) {
        if("daemon_id" in req.body)
        {
            res.json({})
        }
        else
        {
            Daemon.find({ owner: req.user._id }, function(daemon_error, daemons) {
                if(daemon_error)
                {
                    res.status(500).send("Database Error");
                    return;
                }
                var ret_data = [];

                if(daemons.length > 0)
                {
                    async.forEach(daemons, function (daemon, next) {
                        ret_data.push({
                            id: daemon.daemon_id,
                            name: daemon.name,
                            ip: daemon.ip,
                            auth_key: daemon.auth_key,
                            balancer_port_images: daemon.balancer_port_images,
                            balancer_port_child: daemon.balancer_port_child,
                            parent_host: daemon.parent_host,
                            pending_tasks: daemon.pending_tasks,
                            docker_endpoint: daemon.docker_endpoint,
                            offline: false
                        });
                        next();
                    }, function(err){
                        res.json(ret_data);
                    });
                }
                else
                {
                    res.json(ret_data);
                }
            });
        }
    },
    set_task: function(req, res) {
        Daemon.findOne({ $or: [ {daemon_id: req.body.daemon}, {name: req.body.daemon} ], owner: req.user._id }, function(daemon_error, daemon){
            if(daemon_error || !daemon)
            {
                res.status(404).send("Daemon Server not found");
                return;
            }

            var task = new Task();
            task.task_type = req.body.task_type;
            task.data = req.body.data;
            task.daemon = daemon._id;
            task.save(function(){
                res.json({ task_id: task._id.toString(), message: "", error: false });
            });
        });
    },
    task_result: function(req, res) {
        Task.findOne({ _id: req.body.task_id }, function(task_error, task) {
            if(task_error || !task)
            {
                res.status(404).send("Task Not Found");
                return;
            }

            if("end_time" in req.body)
            {
                task.start_time = new Date(req.body.start_time * 1000);
                task.done_time = new Date(req.body.end_time * 1000);
                task.error = req.body.error;
                task.error_message = req.body.message;
                task.done = true;
                task.save(function (err) {
                    if(err)
                    {
                        res.status(500).send("Error Saving");
                        return;
                    }
                    res.json({});
                });
            }
            else
            {
                res.json({
                    error: task.error,
                    message: task.error_message,
                    done: task.done,
                    task_id: task._id.toString()
                });
            }
        });
    },
    register_daemon: function(req, res) {
        Daemon.findOne({ daemon_id: req.daemon_id, owner: req.user._id }, function(daemon_error, d){
            if(daemon_error)
            {
                res.status(500).send("Database error");
                return;
            }

            if(!d)
            {
                d = new Daemon();
                d.daemon_id = req.daemon_id;
                d.owner = req.user._id;
            }

            d.name = req.body.name;
            d.auth_key = req.body.auth_key;
            d.balancer_port_images = req.body.balancer_port_images;
            d.balancer_port_child = req.body.balancer_port_child;
            d.parent_host = req.body.parent_host;
            d.pending_tasks = req.body.pending_tasks;
            d.docker_endpoint = req.body.docker_endpoint;
            d.save(function(err){
                if(err)
                {
                    res.status(500).send("Error saving Daemon to Database");
                    return;
                }
                res.json({});
            });
        });
    },
    notifications: function(req, res) {
        var resp_obj = {
            child_servers: [],
            tasks: [],
            restart: false
        }, log;

        Daemon.findOne({ daemon_id: req.daemon_id, owner: req.user._id }, function(daemon_error, daemon){
            if(daemon_error || !daemon)
            {
                res.status(404).send("Daemon Server not found");
                return;
            }

            log = new StateLogger();
            log.data = req.body.state; // This will contain state from Flaxton Daemon
            log.save(); // This will make async, we don't need to wait until it will be done

            daemon.getTasks(function(task_error, tasks){
                if(task_error)
                {
                    res.status(500).send("Unable to get tasks");
                    return;
                }

                async.forEach(tasks, function(task, next){
                    resp_obj.tasks.push({
                        task_id: task._id.toString(),
                        type: task.task_type,
                        data: task.data,
                        cron: task.cron,
                        start_time: task.start_time,
                        end_time: task.end_time
                    });
                    task.sent = true;
                    task.save(function () {
                        next();
                    });
                }, function(foreach_error){
                    res.json(resp_obj);
                });
            });
        });
    },
    daemon_state: function (req, res) {
        Daemon.find({}, function (daemon_error, daemons) {
            if(daemon_error)
            {
                res.status(500).send("Unable to get Daemon servers");
                return;
            }

            var ret_data = {};
            async.forEach(daemons, function (daemon, next) {
                StateLogger.find({daemon: daemon._id}, null, { sort: { date: -1 } }, function(state_error, loggers){
                    if(daemon_error)
                    {
                        res.status(500).send("Unable to get State loggers for daemon " + daemon.id);
                        return;
                    }
                    ret_data[daemon.id] = loggers[0];
                    console.log(loggers[0]);
                    console.log(typeof loggers[0]);

                    next();
                });
            }, function(end_error){
                if (end_error)
                {
                    if(daemon_error)
                    {
                        res.status(500).send("Error running foreach");
                        return;
                    }
                }
                res.json(ret_data);
            });
        });
    }
};