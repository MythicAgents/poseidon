package listtasks

import (
	// Standard
	"encoding/json"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Listtasks interface {
	Result() map[string]interface{}
}

func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID

	r, err := getAvailableTasks()
	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	byteResult, err := json.MarshalIndent(r.Result(), "", "	")
	msg.UserOutput = string(byteResult)
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
