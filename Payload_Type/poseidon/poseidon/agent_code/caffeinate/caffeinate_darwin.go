//go:build darwin
// +build darwin

package caffeinate

/*
#cgo CFLAGS: -x objective-c -fmacro-backtrace-limit=0 -std=gnu11 -Wobjc-property-no-attribute -Wunguarded-availability-new
#cgo LDFLAGS: -framework Foundation -framework IOKit
#include "caffeinate_wrapper_darwin.h"
*/
import "C"

type CaffeinateRunDarwin struct {
	Successful bool
	Results    string
}

func (j *CaffeinateRunDarwin) Success() bool {
	return j.Successful
}

func (j *CaffeinateRunDarwin) Result() string {
	return j.Results
}

func runCommand(enable bool) (CaffeinateRunDarwin, error) {
	enableInt := 0
	if enable {
		enableInt = 1
	}
	cEnable := C.int(enableInt)
	cresult := C.caffeinate(cEnable)
	result := C.GoString(cresult)
	r := CaffeinateRunDarwin{}
	r.Successful = true
	r.Results = result
	return r, nil
}
