var contact_api_functions = {
	'sms-api': {
		'call': null,
		'creator': function(contact_detail){
			var sender_func = sms.sms_sender_factory(contact_detail['api-name'], contact_detail['props']);
			return function(text, number){
				sender_func(number, text, function(e){
					//TODO: Check if it sent
					console.log(text + " texted to " + number);
				});
			}
		},
		'determinator': function(string){
			var index = string.search(/^\d{10}$/g);
			if(index === 0){
				return true;
			}
			else{
				return false;
			}
		}
	},
	'email-api':{
		'call': null,
		'creator': function(contact_detail){
			var sender_func = email.email_sender_factory(contact_detail['api-name'], contact_detail['props']);
			return function(text, email_addr){
				var subject = contact_detail['subject'];
				var email_text = text + " - " + contact_detail['signature'];
				sender_func(email_addr, subject, email_text, function(e){
					console.log(text + " emailed to " + email_addr);
				});
			}
		},
		'determinator': function(string){
			var index = string.search(/^[A-Z a-z \. \_ \- 1-9]{3,}@[a-z 1-9 \- \_]+\.[a-z]{2,4}$/);
			if(index === 0){
				return true;
			}
			else{
				return false;
			}
		}
	},
	'test-api':{
		'call': null,
		'creator': function(contact_detail){
			return function(text){
				console.log("Logger triggered");
				console.log(text);
			}
		},
		'determinator': function(string){return true;}
	}
}

var check_type = function(string){
	for(var type in contact_api_functions){
		if(contact_api_functions.hasOwnProperty(type)){
			if(contact_api_functions[type]['determinator'](string)){
				return type;
			}
		}
	}
}

exports.initiate = function(contact_apis, check_list){
	for(var i in contact_apis){
		if(contact_apis.hasOwnProperty(i)){
			contact_api_functions[i].call = contact_api_functions[i].creator(contact_apis[i]);
		}
	}

	return function(name, status, callback){ //Sender function
		var text = name + " is " + ((status)?"UP":"DOWN");

		if(check_list.hasOwnProperty(name)){
			contacts = check_list[name]['send_status_to'];
			for(var i = 0; i < contacts.length; i += 1){
				var contact_id = contacts[i];
				var type = check_type(contact_id);

				contact_api_functions[type].call(text, contact_id);
			}
		}
		callback();
	};
}