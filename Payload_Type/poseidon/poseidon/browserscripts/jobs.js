function(task, response){
	let headers = [
            {"plaintext": "kill", "type": "button", "width": 70, "disableSort": true},
			{"plaintext": "command", "type": "string", "width": 200},
            {"plaintext": "params", "type": "string", "fillWidth": true},

        ];
	if(response.length === 0){
		return {"plaintext": "No response yet from agent..."};
	}
	try{
		let data = JSON.parse(response[0]);
		let rows = [];
		for(let j = 0; j < data.length; j++) {
			rows.push({
				"kill": {"button": {
						"name": "",
						"type": "task",
						"ui_feature": "jobs:kill",
						"parameters": data[j]["id"],
						"hoverText": "Kill this job",
						"startIcon": "kill",
					}
				},
				"command": {"plaintext": data[j]["command"]},
				"params": {"plaintext": data[j]["params"]},
			});
		}
		return {"table": [{
			"headers": headers,
			"rows": rows
		}]}
	}catch(error){
		//console.log("error trying to handle list_entitlements browser script", error, response);
		return {"plaintext": response[0]}
	}
}