// +build darwin

package execute_memory

/*
#cgo LDFLAGS: -lm -framework Foundation
#cgo CFLAGS: -Wno-error=implicit-function-declaration
#include "execute_memory_darwin.h"
*/
import "C"

import "strconv"
//import "fmt"

type DarwinExecuteMemory struct {
	Message string
}

func executeMemory(memory []byte, functionName string) (DarwinExecuteMemory, error) {
	res := DarwinExecuteMemory{}
	memoryLength := cap(memory)
	realName := "_main"
	if(functionName != "main" && functionName != ""){
		realName = "__Z" + strconv.Itoa(len(functionName)) + functionName + "v"
	}

	//fmt.Printf("functionName: %s\n", functionName)
	//fmt.Printf("realName: %s\n", realName)
	funcNameMod := C.CString(realName)
	funcName := C.CString("_" + functionName)
	r := C.executeMemory(C.CBytes(memory), C.int(memoryLength), funcName, funcNameMod);
	res.Message = C.GoString(r)
	return res, nil
}
