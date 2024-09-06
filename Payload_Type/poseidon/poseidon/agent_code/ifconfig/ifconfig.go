package ifconfig

import (
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/functions"

	"strings"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// Run - Function that executes
func Run(task structs.Task) {
	msg := task.NewResponse()
	ips := functions.GetCurrentIPAddress()
	msg.UserOutput = strings.Join(ips, "\n")
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
