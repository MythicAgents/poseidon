// +build darwin
// +build arm64

package execute_macho

type DarwinexecuteMacho struct {
	Message string
}

func executeMacho(memory []byte, args []string) (DarwinexecuteMacho, error) {
	res := DarwinexecuteMacho{}
	res.Message = "Not Supported"
	return res, nil
}
