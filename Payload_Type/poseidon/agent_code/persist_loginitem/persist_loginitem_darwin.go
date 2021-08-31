// +build darwin

package persist_loginitem

/*
#cgo CFLAGS: -x objective-c -fmacro-backtrace-limit=0 -std=gnu11 -Wobjc-property-no-attribute -Wunguarded-availability-new
#cgo LDFLAGS: -framework Foundation
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
	cpath := C.CString(path)
	cname := C.CString(name)
	cbool := C.BOOL(global)
	res := C.persist_loginitem(cpath, cname, cbool)
	gores := C.GoBool(res)

	r := PersistLoginItemDarwin{}
	r.Successful = gores
	return r, nil
}
