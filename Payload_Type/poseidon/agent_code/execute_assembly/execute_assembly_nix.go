//go:build linux || darwin
// +build linux darwin

package execute_assembly

import (
	"errors"
)

func executeShellcode(sc []byte, spawnas string) (string, error) {
	return "", errors.New("execute-assembly not compatible with Linux or Mac")
}
