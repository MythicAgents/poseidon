package drives

import (
	// Standard
	"encoding/json"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Drive struct {
	Name             string
	Description      string
	FreeBytes        uint64
	TotalBytes       uint64
	FreeBytesPretty  string
	TotalBytesPretty string
}

func (e Drive) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"name":               e.Name,
		"description":        e.Description,
		"free_bytes":         e.FreeBytes,
		"total_bytes":        e.TotalBytes,
		"free_bytes_pretty":  e.FreeBytesPretty,
		"total_bytes_pretty": e.TotalBytesPretty,
	}
	return json.Marshal(alias)
}

// Run - Function that executes the shell command
func Run(task structs.Task) {
	msg := task.NewResponse()
	res, err := listDrives()
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	driveJson, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	msg.UserOutput = string(driveJson)
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
