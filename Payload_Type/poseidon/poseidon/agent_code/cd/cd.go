package cd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/functions"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// Run - package function to run cd
func Run(task structs.Task) {
	fixedPath := task.Params
	if strings.HasPrefix(fixedPath, "~/") {
		dirname, _ := os.UserHomeDir()
		fixedPath = filepath.Join(dirname, fixedPath[2:])
	}
	fixedPath, _ = filepath.Abs(fixedPath)
	err := os.Chdir(fixedPath)
	msg := task.NewResponse()
	msg.Completed = true
	if err != nil {
		msg.SetError(err.Error())
	} else {
		msg.UserOutput = fmt.Sprintf("changed directory to: %s", task.Params)
		newCwd := functions.GetCwd()
		if newCwd == "" {
			callbackUpdate := structs.CallbackUpdate{Cwd: &task.Params}
			msg.CallbackUpdate = &callbackUpdate
			msg.UserOutput += fmt.Sprintf("\nWarning: Can't list current directory")
		} else {
			callbackUpdate := structs.CallbackUpdate{Cwd: &newCwd}
			msg.CallbackUpdate = &callbackUpdate
		}
	}
	task.Job.SendResponses <- msg
	return
}
