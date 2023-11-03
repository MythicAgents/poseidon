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
	Pid         int    `json:"pid"`
	Data        string `json:"data"`
	UID         int    `json:"uid"`
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
