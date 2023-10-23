package test_password

import (
	"encoding/json"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

var args Arguments

type Arguments struct {
	Username string `json:"username"`
	Password string `json:"password"`
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
