//go:build darwin

package persist_loginitem

/*
#cgo CFLAGS: -x objective-c -fmacro-backtrace-limit=0 -std=gnu11 -Wobjc-property-no-attribute -Wunguarded-availability-new
#cgo LDFLAGS: -framework Foundation -framework CoreServices
#include "persist_loginitem_darwin.h"
*/
import "C"

type PersistLoginItemDarwin struct {
	Message string
}

func runCommand(path string, name string, global bool, list bool, remove bool) PersistLoginItemDarwin {
	if list {
		res := C.listitems()
		return PersistLoginItemDarwin{
			Message: C.GoString(res),
		}
	} else if remove {
		res := C.removeitem(C.CString(path), C.CString(name))
		return PersistLoginItemDarwin{
			Message: C.GoString(res),
		}
	} else if global {
		res := C.addglobalitem(C.CString(path), C.CString(name))
		return PersistLoginItemDarwin{
			Message: C.GoString(res),
		}
	} else {
		res := C.addsessionitem(C.CString(path), C.CString(name))
		return PersistLoginItemDarwin{
			Message: C.GoString(res),
		}
	}
}
