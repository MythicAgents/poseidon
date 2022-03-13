package cd

import (
	// Standard
	"fmt"
	"os"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

//Run - package function to run cd
func Run(task structs.Task) {
	err := os.Chdir(task.Params)
	msg := structs.Response{}
	msg.TaskID = task.TaskID
	if err != nil {
		errResp := structs.Response{}
		errResp.Completed = true
		errResp.TaskID = task.TaskID
		errResp.Status = "error"
		errResp.UserOutput = err.Error()
		task.Job.SendResponses <- errResp
		return
	}

	msg.UserOutput = fmt.Sprintf("changed directory to: %s", task.Params)
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
