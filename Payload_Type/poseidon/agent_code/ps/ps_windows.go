//go:build windows
// +build windows

package ps

import (
	"errors"

	ps "github.com/mitchellh/go-ps"
)

type WindowsProcess struct {
	pid    int
	ppid   int
	binary string
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

func (p *WindowsProcess) ProcessEnvironment() map[string]interface{} {
	var emptyMap map[string]interface{}
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

func Processes() ([]Process, error) {
	//var results []Process
	results := make([]Process, 0, 50)
	procs, err := ps.Processes()
	if err != nil {
		return nil, errors.New("Unable to list windows processes")
	}
	// loop through each process
	for i := 0; i < len(procs); i++ {
		process := &WindowsProcess{
			pid:    procs[i].Pid(),
			ppid:   procs[i].PPid(),
			binary: procs[i].Executable(),
		}
		results = append(results, process)
	}
	return results, nil
}
