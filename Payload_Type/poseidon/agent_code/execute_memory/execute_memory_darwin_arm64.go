// +build arm64

package execute_memory

type DarwinExecuteMemory struct {
	Message string
}

func executeMemory(memory []byte, functionName string, argString string) (DarwinExecuteMemory, error) {
	res := DarwinExecuteMemory{}
	res.Message = "Not Supported"
	return res, nil
}
