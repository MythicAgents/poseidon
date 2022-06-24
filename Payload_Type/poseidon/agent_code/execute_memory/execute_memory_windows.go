// +build windows

package execute_memory

type WindowsExecuteMemory struct {
	Message string
}

func executeMemory(memory []byte, functionName string, argString string) (WindowsExecuteMemory, error) {
	res := WindowsExecuteMemory{}
	res.Message = "Not compatible"
	return res, nil
}
