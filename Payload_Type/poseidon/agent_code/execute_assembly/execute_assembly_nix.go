//go:build linux || darwin
// +build linux darwin

package execute_assembly

import (
	"errors"
)

func executeShellcode(sc []byte) (string, error) {
	return "", errors.New("inject-assembly not compatible with Linux or Mac")
}
