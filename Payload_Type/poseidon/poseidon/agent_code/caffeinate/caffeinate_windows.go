//go:build windows
// +build windows

package caffeinate

import (
	"errors"
)

type CaffeinateRunWindows struct {
	Successful   bool
	Resultstring string
}

func (j *CaffeinateRunWindows) Success() bool {
	return j.Successful
}

func (j *CaffeinateRunWindows) Result() string {
	return j.Resultstring
}

func runCommand(enable bool) (CaffeinateRunWindows, error) {
	n := CaffeinateRunWindows{}
	n.Resultstring = ""
	n.Successful = false
	return n, errors.New("Not implemented")
}
