function(task, responses){
    if(task.status.includes("error")){
        const combined = responses.reduce( (prev, cur) => {
            return prev + cur;
        }, "");
        return {'plaintext': combined};
    }else if(task.completed || task.status === "processed"){
        if(responses.length > 0){
        	let screenshots = [];
        	let errors = [];
        	for(let i = 0; i < responses.length; i++){
        		try{
        			let screenshotData = JSON.parse(responses[i]);
        			screenshots.push(screenshotData.file_id)
				}catch(error){
        			if(responses[i] !== "file downloaded"){
        				errors.push(responses[i]);
					}
				}
			}
        	let responseData = {};
        	if(errors.length > 0){
        		responseData["plaintext"] = "Errors downloading:\n" + JSON.stringify(errors, null, 2);
			}else if(screenshots.length > 0){
        		responseData["screenshot"] = [
					{
						"agent_file_id": screenshots,
						"variant": "contained",
						"name": "View Screenshots"
					}
				]
			}
        	return responseData;
        }else{
            return {"plaintext": "No data to display..."}
        }
    }else{
        // this means we shouldn't have any output
        return {"plaintext": "Not response yet from agent..."}
    }
}