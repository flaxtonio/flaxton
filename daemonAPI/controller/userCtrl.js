/**
 * Created by tigran on 5/15/15.
 */

var mongoose = require("mongoose")
    , md5 = require("MD5")
    , User = mongoose.model("User");

module.exports = {
    register: function(req, res) {
        if(!User.validateUsername(req.body.username))
        {
            res.json({error: true, message: "Invalid username ! Should be only a-zA-Z letters, 0-9 numbers and '_' symbol "});
            return;
        }

        User.find({username: req.body.username}, function(error, users){
            if(error)
            {
                res.json({error: true, message: "Database query error"});
                console.log(error);
                return;
            }
            if(users.length > 0)
            {
                res.json({error: true, message: "User With " + req.body.username + " username is already registered"});
                return;
            }
            var u = new User({username: req.body.username, password: md5(req.body.password)});
            u.save(function(err){
                if(err)
                {
                    res.json({error: true, message: "Error saving user to Database"});
                    return;
                }
                res.json({status: "ok"})
            });
        });
    },
    login: function (req, res) {
        User.findOne({username: req.body.username}, function(error, user){
            if(error)
            {
                res.json({error: true, message: "Database query error"});
                console.log(error);
                return;
            }
            if(user)
            {
                res.json({ authorization: user.encryptForHeader() });
                return;
            }
            res.json({error: true, message: "User not found with username " + req.body.username});
        });
    }
};