// +build windows

package ls

import (
	"encoding/json"
	"os"
)

func GetPermission(finfo os.FileInfo) (string, error) {
	perms := FilePermission{}
	perms.UID = 0
	perms.GID = 0
	perms.Permissions = "rwxrwxrwx"
	perms.User = "n/a"
	perms.Group = "n/a"

	data, err := json.Marshal(perms)
	if err == nil {
		return string(data), nil
	}
	return "", nil
}