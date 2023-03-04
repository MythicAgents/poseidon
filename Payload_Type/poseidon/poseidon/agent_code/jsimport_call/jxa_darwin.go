// +build darwin

package jsimport_call

/*
#cgo CFLAGS: -x objective-c -fmacro-backtrace-limit=0 -std=gnu11 -Wobjc-property-no-attribute -Wunguarded-availability-new
#cgo LDFLAGS: -framework Foundation -framework OSAKit
#include "jxa_wrapper_darwin.h"
*/
import "C"

type JxaRunDarwin struct {
	Successful bool
	Results    string
}

func (j *JxaRunDarwin) Success() bool {
	return j.Successful
}

func (j *JxaRunDarwin) Result() string {
	return j.Results
}

func runCommand(payload string) (JxaRunDarwin, error) {

	cpayload := C.CString(payload)
	cresult := C.runjsimport(cpayload)
	result := C.GoString(cresult)

	r := JxaRunDarwin{}
	r.Successful = true
	r.Results = result
	return r, nil
}
