package cat

import (
	// Standard

	"os"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

//Run - package function to run cat
func Run(task structs.Task) {

	f, err := os.Open(task.Params)

	msg := structs.Response{}
	msg.TaskID = task.TaskID
	if err != nil {

		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	info, err := f.Stat()

	if err != nil {

		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	data := make([]byte, int(info.Size()))
	n, err := f.Read(data)
	if err != nil && n == 0 {

		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	msg.UserOutput = string(data)
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
