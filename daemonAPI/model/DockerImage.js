/**
 * Created by tigran on 5/15/15.
 */

module.exports = function(mongoose){
    var schema = new mongoose.Schema({
        imageName: String,
        imageId: String,
        filename: String,
        uploaded: { type: Date, default: Date.now },
        uploadedBy: { type: mongoose.Schema.Types.ObjectId, ref: 'User' }
    });

    schema.methods.validateDockerImage = function(){
        return true;
    };

    return schema;
};