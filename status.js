var valid_response_codes = [
	200,
	301,
	300,
	302,
	303
];

var match_response_code = function(code, response_codes){
	for(var i = 0, len = response_codes.length; i < len; i += 1){
		if(code.toString() === response_codes[i].toString()){
			return true;
		}
	}
	return false;
}

// exports.check = function(url, callback){
// 	unirest.get(url).end(function (response) {
// 		var up_down_bool = false;
// 		error = response.hasOwnProperty('error')?response['error']:false;
// 		code = response.hasOwnProperty('code')?response['code']:false;
// 		if(error){
// 			up_down_bool = false;
// 		}
// 		else if(code){
// 			up_down_bool = match_response_code(response['code'], valid_response_codes);
// 		}
// 	    callback(up_down_bool);
// 	});
// };

exports.check = function(url, callback){
	http.get(url, function(res){
		var up_down_bool;
		error = res.hasOwnProperty('error')?res['error']:false;
		code = res.hasOwnProperty('statusCode')?res['statusCode']:false;
		if(error){
			up_down_bool = false;
		}
		else if(code){
			up_down_bool = match_response_code(res['statusCode'], valid_response_codes);
		}
		callback(up_down_bool);
	}).on('error', function(err){
		callback(false);
	});
};