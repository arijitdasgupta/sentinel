#!/bin/env nodejs

fs = require('node-fs');
http = require('http');
unirest = require('unirest');
nodemailer = require('nodemailer');
utils = require('./utils.js')
sms =  require('./sms.js');
email = require('./email.js');
logger = require('./logger.js');
status = require('./status.js');
timer = require('./timer.js')

function readConfiguration(callback){
    fs.readFile('./config.json', 'utf-8', function(err, data){
        callback(JSON.parse(data));
    });
}

function initiateHTTPServer(check_list, port){
    var server_port = process.env.OPENSHIFT_NODEJS_PORT || port; //Options for running on OpenShift
    var server_ip_address = process.env.OPENSHIFT_NODEJS_IP || '127.0.0.1'; 

    var server = http.createServer(function (request, response) {
        function send_data(data){
            response.writeHead(200, {"Content-Type": "text/x-json"});
            response.end(JSON.stringify(data));
        }

        //Routine for checking current status of the servers
        var response_data = {};
        var counter = 0;
        var callbacks_recieved = 0;

        for(var i in check_list){
            if(check_list.hasOwnProperty(i)){
                counter += 1;

                status.check(check_list[i]['url'], new function(){
                    var x = i;
                    var recieved = false;
                    return function(bool){
                        response_data[x] = bool?"UP":"DOWN";
                        recieved = true;
                        if(++callbacks_recieved === counter){
                            send_data(response_data);
                        }
                    }
                });
            }
        }

        setTimeout(function(){
            if(callbacks_recieved !== counter){
                send_data(response_data);    
            }            
        }, 30000);
    });

    server.listen(server_port, server_ip_address);
    console.log("Server started...");
}

function initiateTimersWithLogging(check_list, logging_func){
    for(var i in check_list){
        if(check_list.hasOwnProperty(i)){
            timer.create(i, check_list[i], logging_func);
        }
    }
}

function initializer(data){
    var logging_func = logger.initiate(data['contact'], data['check-list']);

    initiateTimersWithLogging(data['check-list'], logging_func);
    initiateHTTPServer(data['check-list'], data['port']);
}

//Starting server
readConfiguration(initializer);