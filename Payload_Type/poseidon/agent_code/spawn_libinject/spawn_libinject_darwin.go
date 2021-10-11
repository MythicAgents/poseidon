// +build darwin

package spawnlibinject

/*
#cgo CFLAGS: -x objective-c -fmacro-backtrace-limit=0 -std=gnu11 -Wobjc-property-no-attribute -Wunguarded-availability-new
#cgo LDFLAGS: -framework Foundation -framework CoreServices
#include "spawn_libinject_darwin.h"
*/
import "C"

type SpawnLibinjectDarwin struct {
	Successful bool
}

func (j *SpawnLibinjectDarwin) Success() bool {
	return j.Successful
}

func runCommand(app string, dylib string, args string, hide bool) (SpawnLibinjectDarwin, error) {
	capp := C.CString(app)
	cdylib := C.CString(dylib)
	cargs := C.CString(args)

	var chide int

	if hide {
		chide = 1
	} else {
		chide = 0
	}

	ihide := C.int(chide)
	res := C.spawn_libinject(capp, cdylib, cargs, ihide)

	r := SpawnLibinjectDarwin{}
	if res == 0 {
		r.Successful = true
	} else {
		r.Successful = false
	}

	return r, nil
}
