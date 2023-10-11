package unsetenv

import (
	// Standard

	"fmt"
	"os"
	"strings"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// Run - interface method that retrieves a process list
func Run(task structs.Task) {
	msg := task.NewResponse()
	params := strings.TrimSpace(task.Params)
	err := os.Unsetenv(params)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	msg.Completed = true
	msg.UserOutput = fmt.Sprintf("Successfully cleared %s", params)
	task.Job.SendResponses <- msg
	return
}
