//Utils
exports.check_properties = function(check_list, object){
	for(var i = 0, len = check_list.length; i < len; i += 1){
		if(!object.hasOwnProperty(check_list[i])){
			return false;
		}
	}
	return true;
}