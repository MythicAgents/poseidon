// +build windows

package drives

import (
	// Standard
	"errors"
)

func listDrives() ([]Drive, error) {
	
	return nil, errors.New("Command not compatible with Windows")
}
