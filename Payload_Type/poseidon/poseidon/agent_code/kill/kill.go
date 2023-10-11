package kill

import (
	// Standard

	"fmt"
	"os"
	"strconv"
	"syscall"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// Run - Function that executes the shell command
func Run(task structs.Task) {
	msg := task.NewResponse()

	pid, err := strconv.Atoi(task.Params)

	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	p, err := os.FindProcess(pid)

	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	p.Signal(syscall.SIGKILL)
	msg.Completed = true
	msg.UserOutput = fmt.Sprintf("Killed process with PID %s", task.Params)
	task.Job.SendResponses <- msg
	return
}
