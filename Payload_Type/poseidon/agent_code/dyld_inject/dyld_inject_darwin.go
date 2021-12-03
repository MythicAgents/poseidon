// +build darwin

package dyldinject

/*
#cgo CFLAGS: -x objective-c -fmacro-backtrace-limit=0 -std=gnu11 -Wobjc-property-no-attribute -Wunguarded-availability-new
#cgo LDFLAGS: -framework Foundation -framework CoreServices
#include "dyld_inject_darwin.h"
*/
import "C"

type DyldInjectDarwin struct {
	Successful bool
}

func (j *DyldInjectDarwin) Success() bool {
	return j.Successful
}

func runCommand(app string, dylib string, hide bool) (DyldInjectDarwin, error) {
	capp := C.CString(app)
	cdylib := C.CString(dylib)

	var chide int

	if hide {
		chide = 1
	} else {
		chide = 0
	}

	ihide := C.int(chide)
	res := C.dyld_inject(capp, cdylib, ihide)

	r := DyldInjectDarwin{}
	if res == 0 {
		r.Successful = true
	} else {
		r.Successful = false
	}

	return r, nil
}
