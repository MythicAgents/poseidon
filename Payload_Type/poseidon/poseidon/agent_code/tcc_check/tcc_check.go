package tcc_check

import (
	"encoding/json"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

var args Arguments

type Arguments struct {
	User string
}

func (e *Arguments) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["user"]; ok {
		e.User = v.(string)
	}
	return nil
}

// Run - package function to run tcc_check
func Run(task structs.Task) {
	msg := task.NewResponse()
	args = Arguments{}
	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	output := checkTCC(args.User)
	msg.UserOutput = output
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
