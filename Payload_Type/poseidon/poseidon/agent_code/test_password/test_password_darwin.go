package test_password

/*
#cgo LDFLAGS: -framework OpenDirectory
#cgo CFLAGS: -Wno-error=implicit-function-declaration
#include "test_password_helper_darwin.h"
*/
import "C"

func testPassword(user string, password string) string {
	return C.GoString(C.testPassword(C.CString(user), C.CString(password)))
}
