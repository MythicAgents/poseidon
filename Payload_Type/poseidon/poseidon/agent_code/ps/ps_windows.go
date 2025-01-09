//go:build windows
// +build windows

package ps

import (
	"errors"
)


func Processes() ([]Process, error) {
	return nil, errors.New("Not Implemented")
}

