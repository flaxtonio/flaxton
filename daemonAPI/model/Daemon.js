/**
 * Created by tigran on 5/24/15.
 */

var async = require("async");

module.exports = function(mongoose){
    var schema = new mongoose.Schema({
        daemon_id: String,
        name: String,
        ip: String,
        auth_key: String,
        /**
         * Key-Value Object for storing current balancing port and images on that port
         *  {
         *      port_number: [
         *              {
         *                  port: port_number,
         *                  image_name: image_name_repo,
         *                  image_port: port_on_image
         *              },
         *              .
         *              .
         *      ],
         *  }
         */
        balancer_port_images: Object,
        /**
         * Key-Value Object for storing current balancing port and child servers on that port
         *  {
         *      port_number: [
         *          {
         *              ip: child_server_ip,
         *              port: port_on_child_server
         *          }
         *      ]
         *  }
         */
        balancer_port_child: Object,
        parent_host: String,
        docker_endpoint: String,
        created: { type: Date, default: Date.now },
        owner: { type: mongoose.Schema.Types.ObjectId, ref: 'User' }
    });

    schema.methods.getTasks = function(cb){
        return this.model("Task").find({ daemon: this._id, sent: false }, cb);
    };

    return schema;
};