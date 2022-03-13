package mv

import (
	// Standard
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	SourceFile      string `json:"source"`
	DestinationFile string `json:"destination"`
}

func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID

	var args Arguments
	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}
	fixedSourcePath := args.SourceFile
	if strings.HasPrefix(fixedSourcePath, "~/") {
		dirname, _ := os.UserHomeDir()
		fixedSourcePath = filepath.Join(dirname, fixedSourcePath[2:])
	}
	args.SourceFile, _ = filepath.Abs(fixedSourcePath)
	fixedDestinationPath := args.DestinationFile
	if strings.HasPrefix(fixedDestinationPath, "~/") {
		dirname, _ := os.UserHomeDir()
		fixedDestinationPath = filepath.Join(dirname, fixedDestinationPath[2:])
	}
	args.DestinationFile, _ = filepath.Abs(fixedDestinationPath)

	if _, err = os.Stat(args.SourceFile); os.IsNotExist(err) {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	err = os.Rename(args.SourceFile, args.DestinationFile)

	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}
	msg.Completed = true
	msg.UserOutput = fmt.Sprintf("Moved %s to %s", args.SourceFile, args.DestinationFile)
	task.Job.SendResponses <- msg
	return
}
