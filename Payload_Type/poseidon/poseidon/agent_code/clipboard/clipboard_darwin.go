//go:build darwin

package clipboard

/*
#cgo LDFLAGS: -framework AppKit -framework Foundation
#cgo CFLAGS: -x objective-c
#include "clipboard_darwin.h"
*/
import "C"
import (
	"unsafe"
)

func GetClipboard(readTypes []string) (string, error) {
	var cArgc C.int = 0
	var cArgv **C.char = nil
	cArgc = C.int(len(readTypes))
	cArgs := make([]*C.char, len(readTypes))
	for i := range cArgs {
		cArgs[i] = nil
	}
	for i, arg := range readTypes {
		cArgs[i] = C.CString(arg)
	}
	cArgv = (**C.char)(unsafe.Pointer(&cArgs[0]))
	contents := C.getClipboard(cArgc, cArgv)
	for i := range cArgs {
		if cArgs[i] != nil {
			defer C.free(unsafe.Pointer(cArgs[i]))
		}
	}
	return C.GoString(contents), nil
}
