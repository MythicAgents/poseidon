package test_password

import (
	"encoding/json"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

var args Arguments

type Arguments struct {
	Username string
	Password string
}

func (e *Arguments) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["username"]; ok {
		e.Username = v.(string)
	}
	if v, ok := alias["password"]; ok {
		e.Password = v.(string)
	}
	return nil
}

// Run - package function to run test_password
func Run(task structs.Task) {
	msg := task.NewResponse()
	args = Arguments{}
	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	output := testPassword(args.Username, args.Password)
	msg.UserOutput = output
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
