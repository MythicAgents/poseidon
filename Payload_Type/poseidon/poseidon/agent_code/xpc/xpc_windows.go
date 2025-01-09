//go:build windows

package xpc

import (
	"errors"
)

func runCommand(command string) ([]byte, error) {
	n := make([]byte, 0)
	return n, errors.New("not implemented")
}
