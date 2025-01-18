//go:build windows

package libinject

import (
	"errors"
)

type WindowsInjection struct {
}

func (l *WindowsInjection) Success() bool {
	return false
}

func injectLibrary(pid int, path string) (WindowsInjection, error) {
	return WindowsInjection{}, errors.New("Not implemented")
}
