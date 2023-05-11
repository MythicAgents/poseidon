//go:build linux
// +build linux

package execute_macho

type LinuxexecuteMacho struct {
	Message string
}

func executeMacho(memory []byte, args []string) (LinuxexecuteMacho, error) {
	res := LinuxexecuteMacho{}
	res.Message = "Not Supported"
	return res, nil
}
