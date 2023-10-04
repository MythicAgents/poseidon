//go:build linux
// +build linux

package pty

import (
	creakPty "github.com/creack/pty"
	"os"
	"os/exec"
	"syscall"
)

func customPtyStart(command *exec.Cmd) (*os.File, error) {
	ptmx, tty, err := creakPty.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tty.Close() }()
	command.Stdin = tty
	command.Stdout = tty
	command.Stderr = tty
	command.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true, // required to get job control
		Setctty: true,
		Ctty:    0,
	}
	err = command.Start()
	if err != nil {
		_ = ptmx.Close()
		return nil, err
	}
	return ptmx, err
}
