// +build linux

package persist_loginitem

import (
	"errors"
)

type PersistLoginItemLinux struct {
	Successful bool
}

func (j *PersistLoginItemLinux) Success() bool {
	return j.Successful
}

func runCommand(name string, path string, global bool) (PersistLoginItemLinux, error) {
	n := PersistLoginItemLinux{}
	n.Successful = false
	return n, errors.New("Not implemented")
}
