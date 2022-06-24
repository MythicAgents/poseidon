// +build windows

package dyldinject

import (
	"errors"
)

type DyldInjectWindows struct {
	Successful bool
}

func (j *DyldInjectWindows) Success() bool {
	return j.Successful
}

func runCommand(app string, dylib string, hide bool) (DyldInjectWindows, error) {
	n := DyldInjectWindows{}
	n.Successful = false
	return n, errors.New("Not compatible")
}
