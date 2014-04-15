exports.create = function(entity_name, check_entity, logging_func){
	var timer = new function(){
		var url = check_entity['url'];
		var interval = check_entity['interval'] * 1000;
		var url_status = true;

		function timer_func(){
			status.check(url, function(bool){
				var current_status = bool;
				if(url_status !== current_status){
					console.log("Something changed", entity_name);
					logging_func(entity_name, current_status, function(){
						console.log("Sent report");
						url_status = current_status;
						setTimeout(timer_func, interval);
					});
				}
				else{
					setTimeout(timer_func, interval);
				}
			});
		}

		timer_func();
	}
}