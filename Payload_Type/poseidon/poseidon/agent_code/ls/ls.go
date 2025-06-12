package ls

import (
	// Standard
	"encoding/json"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/functions"
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
	if finfo.Mode()&os.ModeSetuid != 0 {
		perms.SetUID = true
		if perms.Permissions[3] == 'x' {
			perms.Permissions = perms.Permissions[0:3] + "s" + perms.Permissions[4:]
		} else {
			perms.Permissions = perms.Permissions[0:3] + "S" + perms.Permissions[4:]
		}
	}
	if finfo.Mode()&os.ModeSetgid != 0 {
		perms.SetGID = true
		if perms.Permissions[6] == 'x' {
			perms.Permissions = perms.Permissions[0:6] + "s" + perms.Permissions[7:]
		} else {
			perms.Permissions = perms.Permissions[0:6] + "S" + perms.Permissions[7:]
		}
	}
	if finfo.Mode()&os.ModeSticky != 0 {
		perms.Sticky = true
		perms.Permissions = perms.Permissions[0:8] + "t"
	}
	if finfo.IsDir() {
		perms.Permissions = "d" + perms.Permissions[1:]
	}
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

func ProcessPath(path string) (*structs.FileBrowser, error) {
	var e structs.FileBrowser
	e.SetAsUserOutput = true
	e.Files = make([]structs.FileData, 0)
	fixedPath := path
	if strings.HasPrefix(fixedPath, "~/") {
		dirname, _ := os.UserHomeDir()
		fixedPath = filepath.Join(dirname, fixedPath[2:])
	}
	abspath, _ := filepath.Abs(fixedPath)
	//abspath, _ = filepath.EvalSymlinks(abspath)
	dirInfo, err := os.Stat(abspath)
	filepath.EvalSymlinks(abspath)
	if err != nil {
		return &e, err
	}
	e.IsFile = !dirInfo.IsDir()
	e.Permissions = GetPermission(dirInfo)
	symlinkPath, _ := filepath.EvalSymlinks(abspath)
	if symlinkPath != abspath {
		e.Permissions.Symlink = symlinkPath
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
	e.UpdateDeleted = true
	if dirInfo.IsDir() {
		files, err := os.ReadDir(abspath)
		if err != nil {
			e.Success = false
			e.UpdateDeleted = false
			return &e, err
		}
		fileEntries := make([]structs.FileData, len(files))
		for i := 0; i < len(files); i++ {
			fileEntries[i].IsFile = !files[i].IsDir()
			fileInfo, err := files[i].Info()
			if err != nil {
				fileEntries[i].Permissions = structs.FilePermission{}
				fileEntries[i].FileSize = 0
				fileEntries[i].LastModified = 0
			} else {
				fileEntries[i].Permissions = GetPermission(fileInfo)
				fileEntries[i].FileSize = fileInfo.Size()
				fileEntries[i].LastModified = fileInfo.ModTime().Unix() * 1000
			}
			fileEntries[i].Name = files[i].Name()
			fileEntries[i].FullName = filepath.Join(abspath, files[i].Name())
			symlinkPath, _ = filepath.EvalSymlinks(fileEntries[i].FullName)
			if symlinkPath != fileEntries[i].FullName {
				fileEntries[i].Permissions.Symlink = symlinkPath
			}
			at, err = atime.Stat(fileEntries[i].FullName)
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
		e.UpdateDeleted = false
	}
	return &e, nil
}
func Run(task structs.Task) {
	args := structs.FileBrowserArguments{}
	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		msg := task.NewResponse()
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	if args.Depth == 0 {
		args.Depth = 1
	}
	if args.Host != "" {
		if strings.ToLower(args.Host) != strings.ToLower(functions.GetHostname()) {
			if args.Host != "127.0.0.1" && args.Host != "localhost" {
				msg := task.NewResponse()
				msg.SetError("can't currently list files on remote hosts")
				task.Job.SendResponses <- msg
				return
			}
		}
	}
	var paths = []string{args.Path}
	for args.Depth >= 1 {
		nextPaths := []string{}
		for _, path := range paths {
			msg := task.NewResponse()
			fb, err := ProcessPath(path)
			if err != nil {
				msg.SetError(err.Error())
			}
			msg.FileBrowser = fb
			task.Job.SendResponses <- msg
			if fb == nil {
				continue
			}
			for _, child := range fb.Files {
				if !child.IsFile {
					nextPaths = append(nextPaths, child.FullName)
				}
			}
		}
		paths = nextPaths
		args.Depth--
	}
	msg := task.NewResponse()
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
