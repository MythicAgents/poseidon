//go:build linux && (lsopen || debug)

package lsopen

import (
	"errors"
)

type LSOpenLinux struct {
	Successful bool
}

func (j *LSOpenLinux) Success() bool {
	return j.Successful
}

func runCommand(app string, hide bool, args []string) (LSOpenLinux, error) {
	n := LSOpenLinux{}
	n.Successful = false
	return n, errors.New("Not implemented")
}
