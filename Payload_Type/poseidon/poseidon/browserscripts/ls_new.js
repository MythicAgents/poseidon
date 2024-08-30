function(task, response){
	var rows = [];
	let headers = [
		{"plaintext": "name", "type": "string", "fillWidth": true},
		{"plaintext": "size", "type": "size", "width": 150},
		{"plaintext": "user", "type": "string", "fillWidth": true},
		{"plaintext": "group", "type": "string", "fillWidth": true},
		{"plaintext": "permissions", "type": "string", "width": 150},
		{"plaintext": "modified", "type": "date", "width": 250},
		{"plaintext": "ls", "type": "button", "width": 100},
		{"plaintext": "download", "type": "button", "width": 100},
	];
	let responses = "";
	for(let i = 0; i < response.length; i++){
		responses += response[i];
	}
	let data = JSON.parse(responses);
	let ls_path = "";
	if(data["parent_path"] === "/"){
		ls_path = data["parent_path"] + data["name"];
	}else{
		ls_path = data["parent_path"] + "/" + data["name"];
	}
	let perms = data['permissions'];
	if(data["is_file"]){
		rows.push({
			"name": {"plaintext": data['name'],
				"startIcon": data["is_file"] ? "file":"openFolder",
				"startIconColor": data["is_file"] ? "": "gold",
				"copyIcon": true },
			"size": {"plaintext": data['size']},
			"modified": {"plaintext": (new Date(data["modify_time"])).toISOString(),
				"plaintextHoverText":  (new Date(data["modify_time"])).toDateString()},
			"user": {"plaintext": perms['user']},
			"group": {"plaintext": perms['group']},
			"permissions": {"plaintext": perms["permissions"]},
			"ls": {"button": {
					"name": "",
					"type": "task",
					"ui_feature": "file_browser:list",
					"parameters": ls_path,
					"hoverText": "Issue ls for this entry",
					"startIcon": "list",
				}
			},
			"download": {"button": {
					"name": "",
					"type": "task",
					"ui_feature": "file_browser:download",
					"parameters": ls_path,
					"hoverText": "Download this file",
					"startIcon": "download",
					"disabled": !data["is_file"],
				}
			}
		});
	}

	let files = data['files'];
	for (let j = 0; j < files.length; j++)
	{
		let perms = files[j]['permissions'];
		if(data["parent_path"] === "/"){
			ls_path = data["parent_path"] + data["name"] + "/" + files[j]['name'];
		}else{
			ls_path = data["parent_path"] + "/" + data["name"] + "/" + files[j]['name'];
		}
		rows.push({
			"name": {"plaintext": files[j]['name'], "startIcon": files[j]["is_file"] ? "file":"openFolder",
				"copyIcon": true,
				"startIconColor": files[j]["is_file"] ? "": "gold"
			},
			"size": {"plaintext": files[j]['size']},
			"modified": {"plaintext": (new Date(files[j]["modify_time"])).toISOString(),
				"plaintextHoverText":(new Date(files[j]["modify_time"])).toDateString()},
			"user": {"plaintext": perms['user']},
			"group": {"plaintext": perms['group']},
			"permissions": {"plaintext": perms["permissions"]},
			"ls": {"button": {
					"name": "",
					"type": "task",
					"ui_feature": "file_browser:list",
					"parameters": ls_path,
					"hoverText": "Issue ls for this entry",
					"startIcon": "list",
				}
			},
			"download": {"button": {
					"name": "",
					"type": "task",
					"ui_feature": "file_browser:download",
					"parameters": ls_path,
					"hoverText": "Download this file",
					"startIcon": "download",
					"disabled": !files[j]["is_file"],
				}
			}
		});
	}
	return {"table":[{
			"headers": headers,
			"rows": rows,
		}]};

}