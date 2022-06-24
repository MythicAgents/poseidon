package ls

import (
	// Standard
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	// 3rd Party
	"github.com/djherbis/atime"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type FilePermission struct {
	UID         int    `json:"uid"`
	GID         int    `json:"gid"`
	Permissions string `json:"permissions"`
	User        string `json:"user,omitempty"`
	Group       string `json:"group,omitempty"`
}

func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID
	args := structs.FileBrowserArguments{}
	json.Unmarshal([]byte(task.Params), &args)
	var e structs.FileBrowser
	fixedPath := args.Path
	if strings.HasPrefix(fixedPath, "~/") {
		dirname, _ := os.UserHomeDir()
		fixedPath = filepath.Join(dirname, fixedPath[2:])
	}
	abspath, _ := filepath.Abs(fixedPath)
	dirInfo, err := os.Stat(abspath)
	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}
	e.IsFile = !dirInfo.IsDir()

	e.Permissions.Permissions, err = GetPermission(dirInfo)
	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		e.Success = false
		msg.FileBrowser = &e
		task.Job.SendResponses <- msg
	}
	e.Filename = dirInfo.Name()
	e.ParentPath = filepath.Dir(abspath)
	if strings.Compare(e.ParentPath, e.Filename) == 0 {
		e.ParentPath = ""
	}
	e.FileSize = dirInfo.Size()
	e.LastModified = dirInfo.ModTime().Unix() * 1000
	at, err := atime.Stat(abspath)
	if err != nil {
		e.LastAccess = 0
	} else {
		e.LastAccess = at.Unix() * 1000
	}
	e.Success = true
	if dirInfo.IsDir() {
		files, err := ioutil.ReadDir(abspath)
		if err != nil {
			msg.UserOutput = err.Error()
			msg.Completed = true
			msg.Status = "error"
			e.Success = false
			msg.FileBrowser = &e
			task.Job.SendResponses <- msg
			return
		}

		fileEntries := make([]structs.FileData, len(files))
		for i := 0; i < len(files); i++ {
			fileEntries[i].IsFile = !files[i].IsDir()
			fileEntries[i].Permissions.Permissions, _ = GetPermission(files[i])
			fileEntries[i].Name = files[i].Name()
			fileEntries[i].FullName = filepath.Join(abspath, files[i].Name())
			fileEntries[i].FileSize = files[i].Size()
			fileEntries[i].LastModified = files[i].ModTime().Unix() * 1000
			at, err := atime.Stat(abspath)
			if err != nil {
				fileEntries[i].LastAccess = 0
			} else {
				fileEntries[i].LastAccess = at.Unix() * 1000
			}
		}
		e.Files = fileEntries
	} else {
		fileEntries := make([]structs.FileData, 0)
		e.Files = fileEntries
	}
	msg.Completed = true
	msg.FileBrowser = &e
	temp, _ := json.Marshal(msg.FileBrowser)
	msg.UserOutput = string(temp)
	task.Job.SendResponses <- msg
	return
}
