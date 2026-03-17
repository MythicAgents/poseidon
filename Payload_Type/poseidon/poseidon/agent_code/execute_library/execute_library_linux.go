//go:build linux && (execute_library || debug)

package execute_library

type LinuxExecuteMemory struct {
	Message string
}

func executeLibrary(filePath string, functionName string, args []string) (LinuxExecuteMemory, error) {
	res := LinuxExecuteMemory{}
	res.Message = "Not Supported"
	return res, nil
}
