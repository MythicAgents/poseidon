//go:build linux && (xpc || xpc_load || xpc_manageruid || xpc_procinfo || xpc_send || xpc_service || xpc_submit || xpc_unload || debug)

package xpc

import (
	"errors"
)

func runCommand(command string) ([]byte, error) {
	n := make([]byte, 0)
	return n, errors.New("not implemented")
}
