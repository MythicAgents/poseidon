// +build windows

package ls

import (
	"os"
	"errors"
)

func GetPermission(finfo os.FileInfo) (string, error) {
	return "", errors.New("Not implemented")
}