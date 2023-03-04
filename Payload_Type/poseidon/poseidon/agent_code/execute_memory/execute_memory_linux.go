// +build linux

package execute_memory

type LinuxExecuteMemory struct {
	Message string
}

func executeMemory(memory []byte, functionName string, argString string) (LinuxExecuteMemory, error) {
	res := LinuxExecuteMemory{}
	res.Message = "Not Supported"
	return res, nil
}
