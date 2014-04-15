Sentinel
========

A simple HTTP status monitoring tool with up/down notification via email, sms etc.

Operation
---------

It constantly pings HTTP to the sites/ips of interest and whenever one goes DOWN/UP it notifies the set email/phone number.

It runs on node.

Notifications
-------------

Currently supports notifications via,

1. Gmail (for email)
2. Mashape Site2SMS (for SMS) [API Docs](https://www.mashape.com/blaazetech/site2sms#!) - needs API key from [Mashape](https://www.mashape.com/)


The notifications which are sent via email has configurable signature and subject fields in config.json

Installation
-------------

```
git clone https://github.com/rivalslayer/sentinel
cd sentinel
npm install
cp config.json.sample config.json
```

Edit config.json

If you are running the server in your home/local server, try installing forever and running `server.js` with `forever`

###Installing forever
```
sudo npm install -g forever
```

###Running server with forever
```
forever start server.js
```
###RedHat OpenShift
The code is OpenShift ready. If you want to deploy the monitor to OpenShift, you may edit the configuration and push the code to *RedHat OpenShift*.


Configuration
-------------

```
{
    "server": "localhost",
	"port": 7000,
	"contact":{
		"sms-api":{
			"api-name": "site2sms-mashape",
			"props": {
				"uid": "uid",
				"pwd": "pwd",
				"mashape-auth": "mashape-auth-token"
			},
			"enabled": true
		},
		"email-api":{
			"api-name": "gmail-smtp",
			"props":{
				"uid": "uid@jed-i.in",
				"pwd": "pwd"
			},
			"subject": "Message from the Wall",
			"signature": "Men of Nightswatch"
		}
	},
	"check-list":{
		"Google": {
			"url": "http://google.com",
			"interval": 60,
			"send_status_to": ["email@email.com", "email2@email.com", "9087654321"]
		},
		"Yahoo": {
			"url": "http://yahoo.com",
			"interval": 60,
			"send_status_to": ["email@email.com", "9087654321"]
		}
	}
}
```

Configuration is JSON based (`config.json`). `check-list` is collection of  websites that the monitor is ought to check. `interval` in seconds.

Configure the SMS and email API with your credentials.

1. SMS API `site2sms-mashape`
2. Email API `gmail-smtp`

##And...

I am working on adding more APIs. If any of you is adding another API, do send me a pull request. :-)