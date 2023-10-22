package tcc_check

/*
#cgo LDFLAGS: -framework Foundation
#cgo CFLAGS: -Wno-error=implicit-function-declaration
#include "tcc_check_helper_darwin.h"
*/
import "C"

func checkTCC(user string) string {
	return C.GoString(C.checkTCC(C.CString(user)))
}
