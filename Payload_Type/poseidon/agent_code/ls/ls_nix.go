// +build darwin linux

package ls

import (
	"encoding/json"
	"syscall"
	"os"
	"os/user"
	"strconv"
)

func GetPermission(finfo os.FileInfo) (string, error) {
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
		return string(data), nil
	}
	return "", nil
}