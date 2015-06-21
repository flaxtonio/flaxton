/**
 * Created by tigran on 5/14/15.
 */

var fs = require('fs')
    , mime = require('mime')
    , path = require('path')
    , mongoose = require("mongoose")
    , DockerImage = mongoose.model("DockerImage")
    , Task = mongoose.model("Task")
    , Daemon = mongoose.model("Daemon")
    , async = require("async");

module.exports = {
    index: function(req, res) {
        res.json({ message: 'Flaxton Container API Version 0.1' });
    },

    get_docker_images: function (req, res) {
        if ("image_id" in req.params)
        {
            DockerImage.findOne({ imageId: req.params["image_id"], "uploadedBy": req.user._id }, function (err, docker_image) {
                if(docker_image)
                {
                    var file = __dirname + "/../uploads/" + docker_image.filename
                        , mimeType = mime.lookup(docker_image.filename);
                    res.setHeader('Content-disposition', 'attachment; filename=' + docker_image.filename);
                    res.setHeader('Content-type', mimeType);

                    fs.createReadStream(file).pipe(res);
                }
                else
                {
                    res.json({error: true, message: "Docker Image not Found"});
                }
            });
        }
        else
        {
            DockerImage.find({ "uploadedBy": req.user._id }, function (err, docker_images) {
                var ret_json = {
                    count: docker_images.length,
                    images: []
                };
                async.forEach(docker_images, function (image, callback) {
                    ret_json.images.push({
                        name: image.imageName,
                        id: image.imageId,
                        uploaded: image.uploaded
                    });
                    callback();
                }, function (err) {
                    if(err)
                    {
                        res.json({ error: true, message: "Error parsing trough docker images" });
                    }
                    else
                    {
                        res.send(ret_json);
                    }
                });
            });
        }
    },
    /* Handle Container Image Tar file : User should be authenticated
       {
            image_info: type from Golang lib.TransferContainerCall
       }

       filename should be "docker_image"
     */
    upload_docker_image: function(req, res) {
        var resp_obj = {
            error: false,
            message: "",
            done: false,
            task_id: ""
        };

        if(!("image_info" in req.body))
        {
            resp_obj.error = true;
            resp_obj.message = "image_name and image_id are required";
            res.json(resp_obj);
            return;
        }

        if ("docker_image" in req.files)
        {
            var image_info;
            if(typeof req.body["image_info"] == "string")
            {
                try{
                    image_info = JSON.parse(req.body["image_info"]);
                } catch (e){
                    resp_obj.error = true;
                    resp_obj.message = "Unable to parse image_info to JSON -> " + req.body["image_info"];
                    res.json(resp_obj);
                    return;
                }
            }
            else
            {
                if(typeof req.body["image_info"] == "object")
                {
                    image_info = req.body["image_info"];
                }
                else
                {
                    resp_obj.error = true;
                    resp_obj.message = "image_info should be 'string' or 'object' -> ";
                    res.json(resp_obj);
                    return;
                }
            }

            Daemon.findOne({ $or: [{daemon_id: image_info["destination"]}, {name: image_info["destination"]}, {ip: image_info["destination"]}] }
                , function(daemon_err, daemon){
                    if(daemon_err || !daemon)
                    {
                        resp_obj.error = true;
                        resp_obj.message = "Error in searching daemon with -> " + image_info["destination"];
                        res.json(resp_obj);
                        return;
                    }

                    var tmp_path = req.files["docker_image"].path
                        , target_path = __dirname + '/../uploads/' + image_info["image_id"] + ".tar";

                    fs.rename(tmp_path, target_path, function(err) {
                        if (err)
                        {
                            resp_obj.error = true;
                            resp_obj.message = "Unable to save file";
                            res.json(resp_obj);
                            return;
                        }

                        DockerImage.findOne({ imageId: req.body["image_id"], "uploadedBy.username": req.user.username }).populate("User").exec(function (err, docker_image) {
                            var img;
                            if(docker_image)
                            {
                                img = docker_image;
                            }
                            else
                            {
                                img = new DockerImage();
                            }

                            img.imageName = image_info["image_name"];
                            img.filename = image_info["image_id"] + ".tar";
                            img.uploadedBy = req.user._id;
                            img.imageId = image_info["image_id"];
                            img.save(function (err) {
                                if(err)
                                {
                                    res.json({error: true, message: "Error saving to database"})
                                }
                                else
                                {
                                    var task = new Task();
                                    task.data = image_info;
                                    task.daemon = daemon._id;
                                    task.task_type = "container_transfer";
                                    task.save(function (save_err) {
                                        if(save_err)
                                        {
                                            resp_obj.error = true;
                                            resp_obj.message = "Unable to save task";
                                            res.json(resp_obj);
                                            return;
                                        }

                                        resp_obj.done = false;
                                        resp_obj.error = false;
                                        resp_obj.message = "";
                                        resp_obj.task_id = task._id.toString();
                                        res.json(resp_obj);
                                    });
                                }
                            });
                        });
                    });
            });
        }
        else
        {
            res.json({ status: "error", error: "file not uploaded" });
        }
    }
};