//go:build windows

package drives

import (
	"errors"
)

func listDrives() ([]Drive, error) {
	return nil, errors.New("Not implemented")
}
