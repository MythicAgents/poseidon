//go:build windows
// +build windows

package pty

import (
	"os"
	"os/exec"
	"errors"
)

func customPtyStart(command *exec.Cmd) (*os.File, error) {
	return nil, errors.New("Not Implemented")
}
