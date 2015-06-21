/**
 * Created by tigran on 5/14/15.
 */

var express    = require('express')
    , router = express.Router()
    , dockerCtrl = require("./controller/dockerCtrl")
    , userCtrl = require("./controller/userCtrl")
    , daemonCtrl = require("./controller/daemonCtrl")
    , passport = require("passport")
    , mongoose = require("mongoose")
    , User = mongoose.model("User");

// route middleware to make sure a user is logged in
function isLoggedIn(req, res, next) {

    // if user is authenticated in the session, carry on
    if (req.isAuthenticated())
        return next();

    // This header should contain 'username|md5_password' encrypted using this algorithm and password
    // Or if request is coming from Daemon server it should have encrypted 'username|md5_password'|daemon_id
    if("authorization" in req.headers)
    {
        if(req.headers["authorization"].indexOf("|")) // if it contains "|" symbol then the request from Daemon
        {
            var auth_daemon = req.headers["authorization"].split("|");
            req.daemon_id = auth_daemon[1];
            req.headers["authorization"] = auth_daemon[0];
        }

        var username_pass = User.decryptFromHeader(req.headers["authorization"]).split("|");
        req.body.username = username_pass[0];
        req.body.password = username_pass[1];
        passport.authenticate('local')(req, res, next);
        return;
    }

    res.redirect('/login');
}

// GET Routes
//router.get('/', handlers.index);
router.get('/images', isLoggedIn, dockerCtrl.get_docker_images);
router.get('/images/:image_id', isLoggedIn, dockerCtrl.get_docker_images);


// POST Routes
router.post('/images/add', isLoggedIn, dockerCtrl.upload_docker_image);
router.post('/user/login', passport.authenticate('local'), userCtrl.login);
router.post('/daemon', isLoggedIn, daemonCtrl.register_daemon); // Registering daemon
router.post('/daemon/list', isLoggedIn, daemonCtrl.get_daemon); // Registering daemon
router.post('/notify', isLoggedIn, daemonCtrl.notifications);
router.post('/task', isLoggedIn, daemonCtrl.task_result);
router.post('/task/add', isLoggedIn, daemonCtrl.set_task);

// DockingLogic API calls


module.exports = router;