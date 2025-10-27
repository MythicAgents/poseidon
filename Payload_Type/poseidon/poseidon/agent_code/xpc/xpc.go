package xpc

import (
	// Standard
	"encoding/json"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

var args Arguments

type Arguments struct {
	Command     string
	ServiceName string
	Program     string
	File        string
	Pid         int
	Data        string
	UID         int
}

func (e *Arguments) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["command"]; ok {
		e.Command = v.(string)
	}
	if v, ok := alias["servicename"]; ok {
		e.ServiceName = v.(string)
	}
	if v, ok := alias["program"]; ok {
		e.Program = v.(string)
	}
	if v, ok := alias["file"]; ok {
		e.File = v.(string)
	}
	if v, ok := alias["pid"]; ok {
		e.Pid = int(v.(float64))
	}
	if v, ok := alias["data"]; ok {
		e.Data = v.(string)
	}
	if v, ok := alias["uid"]; ok {
		e.UID = int(v.(float64))
	}
	return nil
}

func Run(task structs.Task) {
	msg := task.NewResponse()
	args = Arguments{}
	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	res, err := runCommand(args.Command)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	msg.UserOutput = string(res)
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
