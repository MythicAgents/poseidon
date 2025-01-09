package persist_launchd

import (

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

func runCommand(task structs.Task) {
	msg := task.NewResponse()
	msg.SetError("Not implemented")
	task.Job.SendResponses <- msg
	return
}
