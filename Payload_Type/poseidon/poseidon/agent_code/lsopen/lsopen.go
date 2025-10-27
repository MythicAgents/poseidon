package lsopen

import (
	// Standard
	"encoding/json"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	Application string
	HideApp     bool
	AppArgs     []string
}

func (e *Arguments) parseStringArray(configArray []interface{}) []string {
	urls := make([]string, len(configArray))
	if configArray != nil {
		for l, p := range configArray {
			urls[l] = p.(string)
		}
	}
	return urls
}
func (e *Arguments) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["application"]; ok {
		e.Application = v.(string)
	}
	if v, ok := alias["hideApp"]; ok {
		e.HideApp = v.(bool)
	}
	if v, ok := alias["appArgs"]; ok {
		e.AppArgs = e.parseStringArray(v.([]interface{}))
	}
	return nil
}

func Run(task structs.Task) {
	msg := task.NewResponse()

	args := Arguments{}
	err := json.Unmarshal([]byte(task.Params), &args)

	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}

	r, err := runCommand(args.Application, args.HideApp, args.AppArgs)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}

	if r.Successful {
		msg.UserOutput = "Successfully spawned application."
		msg.Completed = true
		task.Job.SendResponses <- msg
	} else {
		msg.SetError("Failed to spawn application.")
		task.Job.SendResponses <- msg
	}
	return
}
