//API reference
//Dependency: unirest
var API_references = {
	"site2sms-mashape":{
		check_list: ['uid', 'pwd', 'mashape-auth'],
		sender_factory: function(props){
			if(utils.check_properties(this.check_list, props)){
				return function(number, text, callback){
					var message = escape(text);
					var URL = "https://site2sms.p.mashape.com/index.php?uid={{uid}}&pwd={{pwd}}&phone={{phone}}&msg={{message}}";
					var request = URL.replace("{{phone}}", number).replace("{{message}}", message).replace("{{uid}}", props['uid']).replace("{{pwd}}", props['pwd']);

					var Request = unirest.get(request).headers({ 
					    "X-Mashape-Authorization": props['mashape-auth']
					}).end(function (response) {
					    callback(response);
					});
				}
			}
			else{
				return null;
			}
		}
	}
};

exports.sms_sender_factory = function(api,props){
	return API_references[api].sender_factory(props);
}