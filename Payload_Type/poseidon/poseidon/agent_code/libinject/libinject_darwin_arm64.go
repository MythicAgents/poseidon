//go:build darwin && arm64
// +build darwin,arm64

package libinject
/*
#cgo CFLAGS: -Wno-error=implicit-function-declaration
#cgo LDFLAGS: -framework Foundation -framework Security
#include "libinject_darwin_arm64.h"
*/
import "C"

type DarwinInjection struct {
	Target      int
	Successful  bool
	Payload     []byte
	LibraryPath string
}

func (l *DarwinInjection) TargetPid() int {
	return l.Target
}

func (l *DarwinInjection) Success() bool {
	return l.Successful
}

func (l *DarwinInjection) Shellcode() []byte {
	return l.Payload
}

func (l *DarwinInjection) SharedLib() string {
	return l.LibraryPath
}

func injectLibrary(pid int, path string) (DarwinInjection, error) {
	res := DarwinInjection{}
	i := C.int(pid)
	cpath := C.CString(path)

	r := C.inject(i, cpath)
	res.Successful = true
	if r != 0 {
		res.Successful = false
	}
	return res, nil
}
