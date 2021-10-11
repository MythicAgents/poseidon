// +build linux

package spawnlibinject

import (
	"errors"
)

type SpawnLibinjectLinux struct {
	Successful bool
}

func (j *SpawnLibinjectLinux) Success() bool {
	return j.Successful
}

func runCommand(app string, dylib string, args string, hide int) (SpawnLibinjectLinux, error) {
	n := SpawnLibinjectLinux{}
	n.Successful = false
	return n, errors.New("Not implemented")
}
