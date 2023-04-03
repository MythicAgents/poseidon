// +build darwin
// +build amd64

package execute_macho

//#cgo LDFLAGS: -lm -framework Foundation
//#cgo CFLAGS: -Wno-error=implicit-function-declaration -Wno-deprecated-declarations -Wno-format -Wno-int-conversion
//#include <stdio.h>
//#include <stdlib.h>
//#include "execute_macho_darwin.h"
import "C"
import "os"
import "unsafe"
import "io"
import "bytes"
import "strings"
import "syscall"
import "log"

type DarwinexecuteMacho struct {
	Message string
}

func executeMacho(memory []byte, argString string) (DarwinexecuteMacho, error) {

    res := DarwinexecuteMacho{}

    args := strings.Split(argString, " ")

    // Create our C-esque arguments
    c_argc := C.int(len(args))
    c_argv := C.allocArgv(c_argc)
    defer C.free(c_argv)

    // Convert each argv to a char*
    for i, arg := range args {
        tmp := C.CString(arg)
        defer C.free(unsafe.Pointer(tmp))
        C.addArg(c_argv, tmp, C.int(i))
    }
    cBytes := C.CBytes(memory)
    defer C.free(cBytes)
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

    C.execMachO((*C.char)(cBytes), cLenBytes, c_argc, c_argv)

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
