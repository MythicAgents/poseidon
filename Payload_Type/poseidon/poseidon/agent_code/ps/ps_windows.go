//go:build windows
// +build windows

package ps

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

type WindowsProcess struct {
	pid            int
	ppid           int
	state          rune
	pgrp           int
	sid            int
	architecture   string
	binary         string
	owner          string
	bin_path       string
	additionalInfo map[string]interface{}
}

// Pid returns the process identifier
func (p *WindowsProcess) Pid() int {
	return p.pid
}

// PPid returns the parent process identifier
func (p *WindowsProcess) PPid() int {
	return p.ppid
}

func (p *WindowsProcess) Arch() string {
	return ""
}

// Executable returns the process name
func (p *WindowsProcess) Executable() string {
	return p.binary
}

// Owner returns the username the process belongs to
func (p *WindowsProcess) Owner() string {
	return ""
}

func (p *WindowsProcess) BinPath() string {
	return ""
}

func (p *WindowsProcess) ProcessArguments() []string {
	return []string{""}
}

func (p *WindowsProcess) ProcessEnvironment() map[string]string {
	var emptyMap map[string]string
	return emptyMap
}

func (p *WindowsProcess) SandboxPath() string {
	return ""
}

func (p *WindowsProcess) ScriptingProperties() map[string]interface{} {
	var emptyMap map[string]interface{}
	return emptyMap
}

func (p *WindowsProcess) Name() string {
	return p.binary
}

func (p *WindowsProcess) BundleID() string {
	return ""
}

func (p *WindowsProcess) AdditionalInfo() map[string]interface{} {
	return map[string]interface{}{}
}

func Processes() ([]Process, error) {
	var res []Process

	handle, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return res, err
	}
	defer func() {
		_ = windows.CloseHandle(handle)
	}()

	var procEntry windows.ProcessEntry32
	procEntry.Size = uint32(unsafe.Sizeof(procEntry))
	for err = windows.Process32First(handle, &procEntry); err == nil; err = windows.Process32Next(handle, &procEntry) {
		if procEntry.ProcessID != 0 {
			res = append(res, &WindowsProcess{
				pid:    int(procEntry.ProcessID),
				ppid:   int(procEntry.ParentProcessID),
				binary: windows.UTF16ToString(procEntry.ExeFile[:]),
			})
		}
	}
	return res, nil
}
