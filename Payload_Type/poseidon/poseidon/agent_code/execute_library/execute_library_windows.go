//go:build windows

package execute_library

type WindowsExecuteMemory struct {
	Message string
}

func executeLibrary(filePath string, functionName string, args []string) (WindowsExecuteMemory, error) {
	res := WindowsExecuteMemory{}
	res.Message = "Not Supported"
	return res, nil
}
