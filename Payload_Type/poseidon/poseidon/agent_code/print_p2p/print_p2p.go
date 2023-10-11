package print_p2p

import (
	// Standard

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/p2p"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// Run - package function to run print_tcp
func Run(task structs.Task) {
	msg := task.NewResponse()
	msg.UserOutput = p2p.GetInternalP2PMap()
	msg.Completed = true
	msg.Status = "completed"
	task.Job.SendResponses <- msg
	return
}
