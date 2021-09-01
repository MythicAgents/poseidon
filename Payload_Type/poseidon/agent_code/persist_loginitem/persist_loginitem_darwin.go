// +build darwin

package persist_loginitem

/*
#cgo CFLAGS: -x objective-c -fmacro-backtrace-limit=0 -std=gnu11 -Wobjc-property-no-attribute -Wunguarded-availability-new
#cgo LDFLAGS: -framework Foundation -framework CoreServices
#include "persist_loginitem_darwin.h"
*/
import "C"

type PersistLoginItemDarwin struct {
	Successful bool
}

func (j *PersistLoginItemDarwin) Success() bool {
	return j.Successful
}

func runCommand(path string, name string, global bool) (PersistLoginItemDarwin, error) {
	var glbl int
	cpath := C.CString(path)
	cname := C.CString(name)
	if global {
		glbl = 1
	} else {
		glbl = 0
	}
	iglbl := C.int(glbl)
	res := C.persist_loginitem(cpath, cname, iglbl)

	r := PersistLoginItemDarwin{}
	if res == 1 {
		r.Successful = true
	} else {
		r.Successful = false
	}
	return r, nil
}
