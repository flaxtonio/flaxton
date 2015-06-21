/**
 * Created by tigran on 5/30/15.
 */

module.exports = function(mongoose){
    var schema = new mongoose.Schema({
        image: { type: mongoose.Schema.Types.ObjectId, ref: 'DockerImage' },
        container_id: String,
        name: String,
        command: String,
        status: String,
        created: { type: Date },
        uploadedBy: { type: mongoose.Schema.Types.ObjectId, ref: 'User' }
    });

    schema.methods.getDockerImage = function(callback){
        return this.model("DockerImage").findOne({ daemon: this.image, sent: false }, callback);
    };

    return schema;
};