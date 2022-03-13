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

//Run - Function that executes the copy command
func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID
	args := &Arguments{}
	err := json.Unmarshal([]byte(task.Params), args)
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

	copiedBytes, err := copy(args.SourceFile, args.DestinationFile)

	if err != nil {

		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	msg.Completed = true
	msg.UserOutput = fmt.Sprintf("Copied %d bytes to %s", copiedBytes, args.DestinationFile)
	task.Job.SendResponses <- msg
	return
}
