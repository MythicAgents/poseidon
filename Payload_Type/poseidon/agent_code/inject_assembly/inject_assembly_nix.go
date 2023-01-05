//go:build linux || darwin
// +build linux darwin

package inject_assembly

import (
	"errors"
)

func injectShellcode(sc []byte, spawnas string) (string, error) {
	return "", errors.New("inject-assembly not compatible with Linux or Mac")
}
