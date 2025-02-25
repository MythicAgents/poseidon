package chmod

import (
	// Standard
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
}

// Run - Function that executes the copy command
func Run(task structs.Task) {
	msg := task.NewResponse()
	args := &Arguments{}
	err := json.Unmarshal([]byte(task.Params), args)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	fixedFilePath := args.Path
	if strings.HasPrefix(fixedFilePath, "~/") {
		dirname, _ := os.UserHomeDir()
		fixedFilePath = filepath.Join(dirname, fixedFilePath[2:])
	}
	FullPath, _ := filepath.Abs(fixedFilePath)
	octalValue, err := strconv.ParseInt(args.Mode, 8, 64)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	err = os.Chmod(FullPath, os.FileMode(octalValue))
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	msg.Completed = true
	msg.UserOutput = fmt.Sprintf("Set %s to %s", FullPath, args.Mode)
	task.Job.SendResponses <- msg
	return
}
