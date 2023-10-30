package tail

import (
	"encoding/json"
	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"io"
	"os"
)

type tailArgs struct {
	FilePath string `json:"path"`
	Lines    int    `json:"lines"`
}

// Run - package function to run cat
func Run(task structs.Task) {
	msg := task.NewResponse()
	args := tailArgs{}
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

	var cursor int64 = -1
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

		_, err = fileHandle.Seek(cursor, io.SeekEnd)
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
		if cursor != -1 && (char[0] == '\n') {
			lineCount++
		}

		if cursor == -filesize || lineCount == args.Lines {
			break
		}
		cursor -= 1
	}
	data := make([]byte, -1*cursor)
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
