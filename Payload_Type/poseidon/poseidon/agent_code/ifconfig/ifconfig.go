//go:build (linux || darwin) && (ifconfig || debug)

package ifconfig

import (
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/tasks/taskRegistrar"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/functions"

	"strings"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

func init() {
	taskRegistrar.Register("ifconfig", Run)
}

// Run - Function that executes
func Run(task structs.Task) {
	msg := task.NewResponse()
	ips := functions.GetCurrentIPAddress()
	msg.UserOutput = strings.Join(ips, "\n")
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
