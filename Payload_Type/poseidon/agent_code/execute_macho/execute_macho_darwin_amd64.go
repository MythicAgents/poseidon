// +build darwin
// +build amd64

package execute_macho

//#cgo LDFLAGS: -lm -framework Foundation
//#cgo CFLAGS: -Wno-error=implicit-function-declaration -Wno-deprecated-declarations -Wno-format -Wno-int-conversion
//#include <stdio.h>
//#include <stdlib.h>
//#include "execute_macho_darwin_amd64.h"
import "C"
import "os"
import "unsafe"
import "io"
import "bytes"
import "syscall"
import "log"

type DarwinexecuteMacho struct {
	Message string
}

func executeMacho(memory []byte, args []string) (DarwinexecuteMacho, error) {

    res := DarwinexecuteMacho{}
    var c_argc C.int = 0
    var c_argv **C.char = nil
    
    cBytes := C.CBytes(memory)
    defer func() {
        if cBytes != nil {
            C.free(unsafe.Pointer(cBytes))
        }
    }()
    cLenBytes := C.int(len(memory))

    // Clone Stdout to origStdout.
    origStdout, err := syscall.Dup(syscall.Stdout)
    if err != nil {
        log.Fatal(err)
    }
    // Clone Stdout to origStdout.
    origStderr, err := syscall.Dup(syscall.Stderr)
    if err != nil {
        log.Fatal(err)
    }
    rStdout, wStdout, err := os.Pipe()
    if err != nil {
        log.Fatal(err)
    }
    rStderr, wStderr, err := os.Pipe()
    if err != nil {
        log.Fatal(err)
    }
    reader := io.MultiReader(rStdout, rStderr)

    // Clone the pipe's writer to the actual Stdout descriptor; from this point
    // on, writes to Stdout will go to w.
    if err = syscall.Dup2(int(wStderr.Fd()), syscall.Stdout); err != nil {
        log.Fatal(err)
    }
    // Clone the pipe's writer to the actual Stderr descriptor; from this point
    // on, writes to Stderr will go to w.
    if err = syscall.Dup2(int(wStderr.Fd()), syscall.Stderr); err != nil {
        log.Fatal(err)
    }

    // Background goroutine that drains the reading end of the pipe.
    out := make(chan []byte)
    go func() {
        var b bytes.Buffer
        io.Copy(&b, reader)
        out <- b.Bytes()
    }()
    // END redirect

    if len(args) == 0 {
        // Run default macho without args
        c_argc = C.int(1)
        c_argv = (**C.char)(unsafe.Pointer(&[]*C.char{C.CString(os.Args[0]), nil}[0]))
        C.execMachO((*C.char)(cBytes), cLenBytes, c_argc, c_argv)
    } else {
        // Run macho with argv
        c_argc = C.int(len(args) + 1)
        cArgs := make([](*C.char), len(args)+2)
        for i := range cArgs {
            cArgs[i] = nil
        }
        cArgs[0] = C.CString(os.Args[0])
        for i, arg := range args {
            cArgs[i+1] = C.CString(arg)
        }
        c_argv = (**C.char)(unsafe.Pointer(&cArgs[0]))
        C.execMachO((*C.char)(cBytes), cLenBytes, c_argc, c_argv)
        for i := range cArgs {
            if cArgs[i] != nil {
                defer C.free(unsafe.Pointer(cArgs[i]))
            }
        }
    }

    // BEGIN redirect
    C.fflush(nil)
    wStdout.Close()
    wStderr.Close()
    syscall.Close(syscall.Stdout)
    syscall.Close(syscall.Stderr)
    // Rendezvous with the reading goroutine.
    b := <-out
    // Restore original Stdout and Stderr.
    syscall.Dup2(origStdout, syscall.Stdout)
    syscall.Dup2(origStderr, syscall.Stderr)
    syscall.Close(origStdout)
    syscall.Close(origStderr)
    // END redirect
    res.Message = string(b)
	  return res, nil
}
