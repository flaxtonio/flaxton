/**
 * Created by tigran on 5/30/15.
 */

var async = require("async");

module.exports = function(mongoose){
    var schema = new mongoose.Schema({
        // State data is an object which we will parse later, now we just keeping it to databade
        data: Object,
        time: { type: Date, default: Date.now },
        daemon: { type: mongoose.Schema.Types.ObjectId, ref: 'Daemon' }
    });

    schema.statics.HandleStates = function(containers_info, callback){
        async.forEach(containers_info, function(info, next){

        }, function (err) {

        });
    };

    return schema;
};
