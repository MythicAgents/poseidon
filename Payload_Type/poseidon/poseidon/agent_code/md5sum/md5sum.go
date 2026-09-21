package md5sum

import (
	// Standard
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/tasks/taskRegistrar"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

func init() {
	taskRegistrar.Register("md5sum", Run)
}

// Run - package function to run md5sum
func Run(task structs.Task) {
	msg := task.NewResponse()
	f, err := os.Open(task.Params)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	defer f.Close()
	hashState := md5.New()

	if _, err := io.Copy(hashState, f); err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}

	checkSum := hashState.Sum(nil)

	hashVal := hex.EncodeToString(checkSum)

	msg.UserOutput = fmt.Sprintf("%s %s", hashVal, task.Params)
	msg.Completed = true
	task.Job.SendResponses <- msg
}
