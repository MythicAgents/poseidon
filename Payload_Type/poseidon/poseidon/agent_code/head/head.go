package head

import (
	"encoding/json"
	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"io"
	"os"
)

type Arguments struct {
	FilePath string
	Lines    int
}

func (e *Arguments) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["path"]; ok {
		e.FilePath = v.(string)
	}
	if v, ok := alias["lines"]; ok {
		e.Lines = int(v.(float64))
	}
	return nil
}

// Run - package function to run cat
func Run(task structs.Task) {
	msg := task.NewResponse()
	args := Arguments{}
	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		msg.SetError(fmt.Sprintf("Failed to unmarshal parameters: %s", err.Error()))
		task.Job.SendResponses <- msg
		return
	}
	fileHandle, err := os.Open(args.FilePath)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	defer fileHandle.Close()

	var cursor int64 = 0
	stat, err := fileHandle.Stat()
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	filesize := stat.Size()
	lineCount := 0
	char := make([]byte, 1)
	for {

		_, err = fileHandle.Seek(cursor, io.SeekStart)
		if err != nil {
			msg.SetError(err.Error())
			task.Job.SendResponses <- msg
			return
		}
		_, err = fileHandle.Read(char)
		if err != nil {
			msg.SetError(err.Error())
			task.Job.SendResponses <- msg
			return
		}
		if char[0] == '\n' {
			lineCount++
		}

		if cursor == filesize || lineCount == args.Lines {
			break
		}
		cursor += 1
	}
	data := make([]byte, cursor)
	_, err = fileHandle.Seek(0, io.SeekStart)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	_, err = fileHandle.Read(data)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	msg.UserOutput = string(data)
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
