//go:build darwin

package execute_library

/*
#cgo LDFLAGS: -lm -framework Foundation
#cgo CFLAGS: -Wno-error=implicit-function-declaration
#include "execute_library_darwin.h"
#include <stdio.h>
#include <stdlib.h>
*/
import "C"
import (
	"os"
	"unsafe"
)

//import "fmt"

type DarwinExecuteLibrary struct {
	Message string
}

func executeLibrary(filePath string, functionName string, args []string) (DarwinExecuteLibrary, error) {
	res := DarwinExecuteLibrary{}
	c_argc := C.int(len(args) + 1)
	cArgs := make([]*C.char, len(args)+2)
	for i := range cArgs {
		cArgs[i] = nil
	}
	cArgs[0] = C.CString(os.Args[0])
	for i, arg := range args {
		cArgs[i+1] = C.CString(arg)
	}
	cArgv := (**C.char)(unsafe.Pointer(&cArgs[0]))
	for i := range cArgs {
		if cArgs[i] != nil {
			defer C.free(unsafe.Pointer(cArgs[i]))
		}
	}
	r := C.executeLibrary(C.CString(filePath), C.CString(functionName), c_argc, cArgv)
	res.Message = C.GoString(r)
	return res, nil
}
