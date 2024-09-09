//go:build linux
// +build linux

package caffeinate

import (
	"errors"
)

type CaffeinateRunLinux struct {
	Successful   bool
	Resultstring string
}

func (j *CaffeinateRunLinux) Success() bool {
	return j.Successful
}

func (j *CaffeinateRunLinux) Result() string {
	return j.Resultstring
}

func runCommand(enable bool) (CaffeinateRunLinux, error) {
	n := CaffeinateRunLinux{}
	n.Resultstring = ""
	n.Successful = false
	return n, errors.New("Not implemented")
}
