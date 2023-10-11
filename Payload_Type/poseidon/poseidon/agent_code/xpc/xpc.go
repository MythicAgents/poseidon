package xpc

import (
	// Standard
	"encoding/json"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

var args Arguments

type Arguments struct {
	Command     string `json:"command"`
	ServiceName string `json:"servicename"`
	Program     string `json:"program"`
	File        string `json:"file"`
	KeepAlive   bool   `json:"keepalive"`
	Pid         int    `json:"pid"`
	Data        string `json:"data"`
}

func Run(task structs.Task) {
	msg := task.NewResponse()
	args = Arguments{}
	err := json.Unmarshal([]byte(task.Params), &args)

	if err != nil {

		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	res, err := runCommand(args.Command)

	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	msg.UserOutput = string(res)
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
