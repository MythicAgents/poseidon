// +build windows

package lsopen

import (
	"errors"
)

type LSOpenWindows struct {
	Successful bool
}

func (j *LSOpenWindows) Success() bool {
	return j.Successful
}

func runCommand(app string, hide bool, args []string) (LSOpenWindows, error) {
	n := LSOpenWindows{}
	n.Successful = false
	return n, errors.New("Not implemented")
}
