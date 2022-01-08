package ls

import (
	// Standard
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"

	// 3rd Party
	"github.com/djherbis/atime"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

var mu sync.Mutex


type FilePermission struct {
	UID         int    `json:"uid"`
	GID         int    `json:"gid"`
	Permissions string `json:"permissions"`
	User        string `json:"user,omitempty"`
	Group       string `json:"group,omitempty"`
}

func GetPermission(finfo os.FileInfo) string {
	perms := FilePermission{}
	perms.Permissions = finfo.Mode().Perm().String()
	systat := finfo.Sys().(*syscall.Stat_t)
	if systat != nil {
		perms.UID = int(systat.Uid)
		perms.GID = int(systat.Gid)
		tmpUser, err := user.LookupId(strconv.Itoa(perms.UID))
		if err == nil {
			perms.User = tmpUser.Username
		}
		tmpGroup, err := user.LookupGroupId(strconv.Itoa(perms.GID))
		if err == nil {
			perms.Group = tmpGroup.Name
		}
	}
	data, err := json.Marshal(perms)
	if err == nil {
		return string(data)
	}
	return ""
}

func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID
	args := structs.FileBrowserArguments{}
	json.Unmarshal([]byte(task.Params), &args)
	var e structs.FileBrowser
	abspath, _ := filepath.Abs(args.Path)
	dirInfo, err := os.Stat(abspath)
	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		resp, _ := json.Marshal(msg)
		mu.Lock()
		profiles.TaskResponses = append(profiles.TaskResponses, resp)
		mu.Unlock()
		return
	}
	e.IsFile = !dirInfo.IsDir()

	e.Permissions.Permissions = GetPermission(dirInfo)
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
			msg.FileBrowser = e
			resp, _ := json.Marshal(msg)
			mu.Lock()
			profiles.TaskResponses = append(profiles.TaskResponses, resp)
			mu.Unlock()
			return
		}

		fileEntries := make([]structs.FileData, len(files))
		for i := 0; i < len(files); i++ {
			fileEntries[i].IsFile = !files[i].IsDir()
			fileEntries[i].Permissions.Permissions = GetPermission(files[i])
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
	}
	msg.Completed = true
	msg.FileBrowser = e
	temp, _ := json.Marshal(msg.FileBrowser)
	msg.UserOutput = string(temp)
	resp, _ := json.Marshal(msg)
	mu.Lock()
	profiles.TaskResponses = append(profiles.TaskResponses, resp)
	mu.Unlock()
	return
}
