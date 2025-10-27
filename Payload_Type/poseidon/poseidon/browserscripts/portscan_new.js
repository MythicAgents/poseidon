function(task, response){
	let tables = [];
	let headers = [
			{"plaintext": "ips", "type": "button", "width": 80, "disableSort": true},
            {"plaintext": "hostname", "type": "string", "fillWidth": true, "disableSort": true},
			{"plaintext": "open ports", "type": "string", "fillWidth": true,"disableSort": true}

        ];
	if(response.length === 0){
		return {"plaintext": "No response yet from agent..."};
	}
	try{
		for(let i = 0; i < response.length; i++){
			let data = JSON.parse(response[i]);
			for(let j = 0; j < data.length; j++){
				let rows = [];
				for(let k = 0; k < data[j]["hosts"].length; k++){
					if(data[j]["hosts"][k]["open_ports"] === null){continue}
					rows.push({
						"hostname": {"plaintext": data[j]["hosts"][k]['hostname']},
						"ips": {
							"button": {
								"name": "View IPs",
								"type": "string",
								"value": data[j]["hosts"][k]["pretty_name"],
								"hoverText": "View all IPs",
								"title": "All IPs associated with this host",
							}
						},
						"pretty name": {"plaintext":data[j]["hosts"][k]["pretty_name"]},
						"open ports": {"plaintext": JSON.stringify(data[j]["hosts"][k]["open_ports"])}
					});
				}
				tables.push({
					"title": "Range: " + data[j]["range"],
					"headers": headers,
					"rows": rows
				})
			}

		}
		return {"table":tables};
	}catch(error){
		//console.log("error trying to handle list_entitlements browser script", error, response);
		return {"plaintext": response[0]}
	}
}