package cp

import (
	// Standard
	"encoding/json"
	"fmt"
	"io"
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

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
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
	fixedSourcePath := args.SourceFile
	if strings.HasPrefix(fixedSourcePath, "~/") {
		dirname, err := os.UserHomeDir()
		if err != nil {
			msg.SetError(err.Error())
			task.Job.SendResponses <- msg
			return
		}
		fixedSourcePath = filepath.Join(dirname, fixedSourcePath[2:])
	}
	args.SourceFile, err = filepath.Abs(fixedSourcePath)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	fixedDestinationPath := args.DestinationFile
	if strings.HasPrefix(fixedDestinationPath, "~/") {
		dirname, err := os.UserHomeDir()
		if err != nil {
			msg.SetError(err.Error())
			task.Job.SendResponses <- msg
			return
		}
		fixedDestinationPath = filepath.Join(dirname, fixedDestinationPath[2:])
	}
	args.DestinationFile, err = filepath.Abs(fixedDestinationPath)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	copiedBytes, err := copy(args.SourceFile, args.DestinationFile)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	msg.Completed = true
	msg.UserOutput = fmt.Sprintf("Copied %d bytes to %s", copiedBytes, args.DestinationFile)
	task.Job.SendResponses <- msg
	return
}
