function(task, response){
	var rows = [];
	let all_interesting = ["com.apple.security.cs.allow-jit",
		"com.apple.security.cs.allow-unsigned-executable-memory",
		"com.apple.security.cs.allow-dyld-environment-variables",
		"com.apple.security.cs.disable-library-validation",
		"com.apple.security.cs.disable-executable-page-protection",
		"com.apple.security.cs.debugger", "No Entitlements", "Invalid argument"];
	let headers = [
		{"plaintext": "actions", "type": "button", "width": 100, "disableSort": true},
		{"plaintext": "pid", "type": "number", "width": 100},
		{"plaintext": "code_sign", "type": "number", "width": 150},
		{"plaintext": "interesting", "type": "button", "width": 150, "disableSort": true},
		{"plaintext": "name", "type": "string", "fillWidth": true},

	];
	if(response.length === 0){
		return {"plaintext": "No response yet from agent..."};
	}
	try{
		let responses = "";
		for(let i = 0; i < response.length; i++){
			responses += response[i];
		}
		let permissions = JSON.parse(responses);
		for(let i = 0; i < permissions.length; i++){
			let data = permissions[i];
			let perms = data["entitlements"];
			let interesting = {};
			try{
				for(let j = 0; j < all_interesting.length; j++){
					if(all_interesting[j] in perms){
						interesting[all_interesting[j]] = perms[all_interesting[j]];
					}
				}
			}catch(error){
				console.log("error in list_entitlements browser script", error);
				continue;
			}
			rows.push({
				"pid": {"plaintext": data['process_id']},
				"name": {"plaintext": data['name']},
				"code_sign": {"plaintext": "0x" + data['code_sign'].toString(16)},
				"interesting": {"button":{
						"name": Object.keys(interesting).length + " interesting",
						"value": interesting,
						"disabled": Object.keys(interesting).length === 0,
						"type": "dictionary",
						"leftColumnTitle": "Entitlement",
						"rightColumnTitle": "Values",
						"title": "Viewing Interesting Entitlements"
					}},
				"actions": {"button": {
						"name": "Actions",
						"type": "menu",
						"value": [
							{
								"name": "View Entitlements",
								"type": "dictionary",
								"value": perms,
								"leftColumnTitle": "Entitlement",
								"rightColumnTitle": "Values",
								"title": "Viewing Entitlements"
							},
							{
								"name": "LS Path",
								"type": "task",
								"ui_feature": "file_browser:list",
								"parameters": data["bin_path"]
							},
							{
								"name": "Download File",
								"type": "task",
								"ui_feature": "file_browser:download",
								"parameters": data["bin_path"],
								"startIcon": "download"
							}
						]
					}}
			});
		}
		return {"table":[{
				"headers": headers,
				"rows": rows,
			}],
			"plaintext": "Searched for the following interesting entitlements:\n" + JSON.stringify(all_interesting, null, 2)};
	}catch(error){
		console.log("error trying to handle list_entitlements browser script", error, response);
		let responses = "";
		for(let i = 0; i < response.length; i++){
			responses += response[i];
		}
		return {"plaintext": responses}
	}
}