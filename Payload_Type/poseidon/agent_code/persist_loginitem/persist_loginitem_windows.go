// +build windows

package persist_loginitem

import (
	"errors"
)

type PersistLoginItemWindows struct {
	Successful bool
}

func (j *PersistLoginItemWindows) Success() bool {
	return j.Successful
}

func runCommand(name string, path string, global bool) (PersistLoginItemWindows, error) {
	n := PersistLoginItemWindows{}
	n.Successful = false
	return n, errors.New("Not implemented")
}
