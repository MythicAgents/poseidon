//go:build listtasks || debug

package listtasks

import (
	// Standard
	"encoding/json"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/tasks/taskRegistrar"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

func init() {
	taskRegistrar.Register("listtasks", Run)
}

type Listtasks interface {
	Result() map[string]interface{}
}

func Run(task structs.Task) {
	msg := task.NewResponse()
	r, err := getAvailableTasks()
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	byteResult, err := json.MarshalIndent(r.Result(), "", "	")
	msg.UserOutput = string(byteResult)
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
