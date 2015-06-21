/**
 * Created by tigran on 5/14/15.
 */

// Fist of all lets defined modules for using it from other parts of code
var mongoose = require('mongoose');
require('mongoose.models.autoload')(mongoose, require('path').join(__dirname, '../', 'model'), true).connect('mongodb://127.0.0.1/flaxton');

// call the packages we need
var express    = require('express')
    , app        = express()
    , server = require('http').Server(app)
    , io = require('socket.io')(server)
    , bodyParser = require('body-parser')
    , expressSession = require('express-session')
    , passport = require("passport")
    , LocalStrategy = require('passport-local').Strategy
    , routes = require("./routes")
    , methodOverride = require("method-override")
    , multipart = require('connect-multiparty');


app.use('/', function(req, res, next) {
    res.header("Access-Control-Allow-Origin", "*");
    res.header("Access-Control-Allow-Headers", "X-Requested-With");
    next();
});

app.use(bodyParser.urlencoded({ extended: true, keepExtensions:true }));
app.use(bodyParser.json());
app.use(expressSession({secret: '$G3@GQkE2quZ'}));
app.use(passport.initialize());
app.use(passport.session({ secret: '*6UwtTYSycE@zCN' }));
app.use(methodOverride());
app.use(multipart());

// Global Docker Containers
// TODO: Needs to be removed
Docker_containers = {};

io.on('connection', function(socket){
    socket.on('containers', function (data) {
        socket.emit('containers_list', Docker_containers);
    });
});


app.use (function (error, req, res, next){
    res.status(500).send("Internal Server Error");
});


var User = mongoose.model("User");

passport.use(new LocalStrategy(
    function(username, password, done) {
        User.findOne({ username: username }, function (err, user) {
            if (err) { return done(err); }
            if (!user) {
                return done(null, false, { message: 'Incorrect username.' });
            }
            if (!user.validPassword(password)) {
                return done(null, false, { message: 'Incorrect password.' });
            }
            return done(null, user);
        });
    }
));

passport.serializeUser(function(user, done) {
    done(null, user.id);
});

passport.deserializeUser(function(id, done) {
    User.findById(id, function(err, user) {
        done(err, user);
    });
});

var port = process.env.PORT || 8080;        // set our port

app.use('/', routes);

server.listen(port);
console.log('Server Started on port ' + port);