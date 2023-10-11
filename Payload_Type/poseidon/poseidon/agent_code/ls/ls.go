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
	"syscall"

	// 3rd Party
	"github.com/djherbis/atime"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

func GetPermission(finfo os.FileInfo) structs.FilePermission {
	perms := structs.FilePermission{}
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
	return perms
}

func Run(task structs.Task) {
	msg := task.NewResponse()
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

	e.Permissions = GetPermission(dirInfo)
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
	e.UpdateDeleted = true
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
			fileEntries[i].Permissions = GetPermission(files[i])
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
