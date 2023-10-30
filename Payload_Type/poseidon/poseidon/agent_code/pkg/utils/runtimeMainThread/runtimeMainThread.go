package runtimeMainThread

import (
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"runtime"
)

// https://github.com/golang/go/wiki/LockOSThread
// Arrange that main.main runs on main thread.
func init() {
	runtime.LockOSThread()
}

// Main runs the main SDL service loop.
// The binary's main.main must call sdl.Main() to run this loop.
// Main does not return. If the binary needs to do other work, it
// must do it in separate goroutines.
func Main() {
	for f := range mainfunc {
		f()
	}
}

// queue of work to run in main thread.
var mainfunc = make(chan func())

// DoOnMainThread runs f on the main thread.
func DoOnMainThread(f func(t structs.Task), t structs.Task) {
	done := make(chan bool, 1)
	mainfunc <- func() {
		f(t)
		done <- true
	}
	<-done
}
