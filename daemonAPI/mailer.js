/**
 * Created by tigran on 5/28/15.
 */

// THIS FILE SHOULD BE REMOVED FROM HERE

var nodemailer = require('nodemailer');


module.exports = {
    send: function(content) {
        // create reusable transporter object using SMTP transport
        var transporter = nodemailer.createTransport({
            service: 'Yandex',
            auth: {
                user: 'hello@flaxton.io',
                pass: 'uchecked'
            },
            tls: {
                rejectUnauthorized: false
            }
        });

// setup e-mail data with unicode symbols
        var mailOptions = {
            from: 'hello@flaxton.io', // sender address
            to: 'tigran@flaxton.io', // list of receivers
            subject: 'Web Contact', // Subject line
            html: content // plaintext body
        };

// send mail with defined transport object
        transporter.sendMail(mailOptions, function(error, info){
            if(error){
                return console.log(error);
            }
            console.log('Message sent: ' + info.response);

        });
    }
};