// +build linux

package dyldinject

import (
	"errors"
)

type DyldInjectLinux struct {
	Successful bool
}

func (j *DyldInjectLinux) Success() bool {
	return j.Successful
}

func runCommand(app string, dylib string, hide int) (DyldInjectLinux, error) {
	n := DyldInjectLinux{}
	n.Successful = false
	return n, errors.New("Not implemented")
}
