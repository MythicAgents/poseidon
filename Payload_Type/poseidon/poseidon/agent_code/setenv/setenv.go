package setenv

import (
	// Standard

	"fmt"
	"os"
	"strings"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// Run - Function that executes the shell command
func Run(task structs.Task) {
	msg := task.NewResponse()
	if task.Params == "" {
		msg.SetError("No environment variable given to set. Must be of format:\nsetenv NAME VALUE")
		task.Job.SendResponses <- msg
		return
	}
	if !strings.Contains(task.Params, " ") {
		msg.SetError("Improper command format given. Must be of format:\nsetenv NAME VALUE")
		task.Job.SendResponses <- msg
		return
	}
	parts := strings.SplitAfterN(task.Params, " ", 2)
	parts[0] = strings.TrimSpace(parts[0])
	parts[1] = strings.TrimSpace(parts[1])

	err := os.Setenv(parts[0], parts[1])
	if err != nil {
		msg.UserOutput = err.Error()
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	msg.Completed = true
	msg.UserOutput = fmt.Sprintf("Set %s=%s", parts[0], parts[1])
	task.Job.SendResponses <- msg
	return
}
