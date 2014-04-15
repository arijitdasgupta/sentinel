//API reference
//Dependency: nodemailer
var API_references = {
	"gmail-smtp":{
		check_list: ['uid', 'pwd'],
		sender_factory: function(props){
			if(utils.check_properties(this.check_list, props)){
				var transport = nodemailer.createTransport("SMTP",{
				    service: "Gmail",
				    auth: {
				        user: props['uid'],
				        pass: props['pwd']
				    }
				});
				return function(email_address, subject, text, callback){
					var mailOptions = {
					    from: props['uid'], // sender address
					    to: email_address, // list of receivers
					    subject: subject, // Subject line
					    text: text, // plaintext body
					}
					transport.sendMail(mailOptions, callback);
				}
			}
			else{
				return null;
			}
		}
	}
};

exports.email_sender_factory = function(api, props){
	return API_references[api].sender_factory(props);
}