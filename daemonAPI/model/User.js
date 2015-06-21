/**
 * Created by tigran on 5/15/15.
 */

var md5 = require('MD5')
    , usernameRegex = /^[a-zA-Z0-9_]+$/
    , crypto = require('crypto')
    , algorithm = 'aes-256-ctr'
    , password = 'JT*nc~~u@eXud}n';

module.exports = function(mongoose){

    var schema = new mongoose.Schema({
        username: String,
        password: String,
        lastAccess: { type: Date, default: Date.now },
        ipAddresses: [String]
    });

    schema.methods.validPassword = function(pwd){
        return (md5(pwd) === this.password || pwd === this.password);
    };

    schema.statics.validateUsername = function (username, password) {
        return usernameRegex.test(username);
    };

    schema.methods.encryptForHeader = function () {
        var cipher = crypto.createCipher(algorithm,password);
        var crypted = cipher.update(this.username + "|" + this.password,'utf8','hex');
        crypted += cipher.final('hex');
        return crypted;
    };

    schema.statics.decryptFromHeader = function (authorization) {
        var decipher = crypto.createDecipher(algorithm, password);
        var dec = decipher.update(authorization, 'hex','utf8');
        dec += decipher.final('utf8');
        return dec;
    };

    return schema;
};