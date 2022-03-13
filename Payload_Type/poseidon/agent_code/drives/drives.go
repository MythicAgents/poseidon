package drives

import (
	// Standard
	"encoding/json"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Drive struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	FreeBytes        uint64 `json:"free_bytes"`
	TotalBytes       uint64 `json:"total_bytes"`
	FreeBytesPretty  string `json:"free_bytes_pretty"`
	TotalBytesPretty string `json:"total_bytes_pretty"`
}

//Run - Function that executes the shell command
func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID

	res, err := listDrives()

	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	driveJson, err := json.MarshalIndent(res, "", "    ")
	msg.UserOutput = string(driveJson)
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
