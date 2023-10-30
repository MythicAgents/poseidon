package prompt

/*
#cgo LDFLAGS: -framework Foundation -framework AppKit -framework Security -framework Cocoa
#cgo CFLAGS: -Wno-error=implicit-function-declaration
#include "prompt_darwin.h"
*/
import "C"

func prompt(args Arguments) string {
	cTitle := C.CString(args.TitleText)
	cMessage := C.CString(args.MessageText)
	cIcon := C.CString(args.Icon)
	cMaxTries := C.int(args.MaxTries)
	result := C.GoString(C.prompt(cIcon, cTitle, cMessage, cMaxTries))
	return result
}
