function(task, response){
	let headers = [
            {"plaintext": "host", "type": "string", "fillWidth": true, "disableSort": true},
			{"plaintext": "username", "type": "string", "fillWidth": true},
            {"plaintext": "secret", "type": "string", "fillWidth": true},
			{"plaintext": "status", "type": "string", "width": 200},
			{"plaintext": "copy_status", "type": "string", "width": 200},
			{"plaintext": "output", "type": "string", "fillWidth": true},

        ];
	if(response.length === 0){
		return {"plaintext": "No response yet from agent..."};
	}
	try{
		let data = JSON.parse(response[0]);
		let rows = [];
		for(let j = 0; j < data.length; j++) {
			rows.push({
				"host": {"plaintext": data[j]["host"]},
				"username": {"plaintext": data[j]["username"]},
				"secret": {"plaintext": data[j]["secret"]},
				"status": {"plaintext": data[j]["status"]},
				"output": {"plaintext": data[j]["output"]},
				"copy_status": {"plaintext": data[j]["copy_status"]},
				"rowStyle": {backgroundColor: data[j]["success"] ?  "green" : ""},
			});
		}
		return {"table": [{
			"title": "SSH Auth Output",
			"headers": headers,
			"rows": rows
		}]}
	}catch(error){
		//console.log("error trying to handle list_entitlements browser script", error, response);
		return {"plaintext": response[0]}
	}
}