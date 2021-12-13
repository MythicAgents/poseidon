function(task, response){
	var rows = [];
	let headers = [
            {"plaintext": "name", "type": "string"},
            {"plaintext": "size", "type": "size"},
            {"plaintext": "user", "type": "string"},
            {"plaintext": "group", "type": "string"},
            {"plaintext": "permissions", "type": "string", "width": 8},
			{"plaintext": "modified", "type": "string"},
            {"plaintext": "ls", "type": "button", "width": 10},
        ];

	for(let i = 0; i < response.length; i++){
		let data = JSON.parse(response[i]);
		let ls_path = "";
		if(data["parent_path"] === "/"){
			ls_path = data["parent_path"] + data["name"];
		}else{
			ls_path = data["parent_path"] + "/" + data["name"];
		}
		let perms = JSON.parse(data['permissions']['permissions']);
		rows.push({
			"name": {"plaintext": data['name']},
			"size": {"plaintext": data['size']},
			"modified": {"plaintext": data["modify_time"]},
			"user": {"plaintext": perms['user']},
			"group": {"plaintext": perms['group']},
			"permissions": {"plaintext": perms["permissions"]},
			"ls": {"button": {
					"name": "ls .",
					"type": "task",
					"ui_feature": "file_browser:list",
					"parameters": ls_path

				}
			}
		});
		let files = data['files'];
		for (let j = 0; j < files.length; j++)
		{
			let perms = JSON.parse(files[j]['permissions']['permissions']);
			rows.push({
				"name": {"plaintext": files[j]['name']},
				"size": {"plaintext": files[j]['size']},
				"modified": {"plaintext": files[j]["modify_time"]},
				"user": {"plaintext": perms['user']},
				"group": {"plaintext": perms['group']},
				"permissions": {"plaintext": perms["permissions"]},
				"ls": {"button": {
						"name": "ls .",
						"type": "task",
						"ui_feature": "file_browser:list",
						"parameters": ls_path

					}
				}
			});
		}
		return {"table":[{
            "headers": headers,
            "rows": rows,
            "title": "File Listing Data"
        }]};
	}
}


/*
httpGetAsync("{{http}}://{{links.server_ip}}:{{links.server_port}}{{links.api_base}}/filebrowserobj/" + data['id'] + "/permissions", (response)=>{
                  try{
                      let perms = JSON.parse(response);
                      if(perms['status'] === "success"){
                          try {
                              this.file_browser_permissions = JSON.parse(perms['permissions']);
                          } catch (error) {
                              this.file_browser_permissions = {"Permissions": perms['permissions']};
                          }
                      }else{
                          alertTop("warning", "Failed to fetch permissions: " + perms['error']);
                      }
                  }catch(error){
                      console.log(error);
                      alertTop("danger", "Session expired, please refresh");
                  }
              }, "GET", null);

*/
