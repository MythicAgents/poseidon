function(task, response){
	let rows = [];
	let headers = [
            {"plaintext": "ppid", "type": "number", "width": 100},
            {"plaintext": "pid", "type": "number", "width": 100},
            {"plaintext": "arch", "type": "string", "width": 100},
			{"plaintext": "name", "type": "string", "fillWidth": true},
			{"plaintext": "user", "type": "string", "fillWidth": true},
            {"plaintext": "more", "type": "button", "width": 100, "disableSort": true},
        ];
	if(response.length === 0){
		return {"plaintext": "No response yet from agent..."};
	}
	try {
		let responses = "";
		for(let i = 0; i < response.length; i++){
			responses += response[i];
		}
		let data = JSON.parse(responses);
		for (let j = 0; j < data.length; j++) {
			rows.push({
				"ppid": {"plaintext": data[j]['parent_process_id']},
				"pid": {"plaintext": data[j]['process_id']},
				"arch": {"plaintext": data[j]["architecture"]},
				"name": {"plaintext": data[j]['name']},
				"user": {"plaintext": data[j]['user']},
				"more": {
					"button": {
						"name": "",
						"type": "dictionary",
						"value": {
							"bin_path": data[j]["bin_path"],
							"args": data[j]["args"],
							"env": data[j]["env"],
							"sandboxpath": data[j]["sandboxpath"],
							"scripting properties": data[j]["scripting_properties"],
							"bundle ID": data[j]["bundleid"],
							"additional_info": data[j]["additional_info"],
						},
						"hoverText": "view data for this entry",
						"startIcon": "list",
					}
				}
			});
		}


		return {
			"table": [{
				"headers": headers,
				"rows": rows,
			}]
		};
	}catch(error){
		//console.log("error trying to handle list_entitlements browser script", error, response);
		let responses = "";
		for(let i = 0; i < response.length; i++){
			responses += response[i];
		}
		return {"plaintext": responses}
	}
}