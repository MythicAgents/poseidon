function(task, responses){
  if(task.status === 'error'){return "<pre> Error: Untoggle swtich to see error message(s) </pre>"; }
  let output = "";
  for(let i = 0; i < responses.length; i+=2){
  	if( i+1 < responses.length){
  		//only want to do this if the next response exists, i.e. file_downloaded
  		let status = JSON.parse(responses[i]['response']);
    	let id = status['agent_file_id'];
    	output += "<img src='/api/v1.4/files/screencaptures/" + id + "' width='100%'/>"
  	}else{
  		output += "<pre> downloading pieces ...</pre>";
  	}
  }
  return output;
}
