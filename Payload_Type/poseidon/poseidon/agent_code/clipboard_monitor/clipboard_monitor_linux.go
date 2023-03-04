//go:build linux
// +build linux

package clipboard_monitor

import (
	"errors"
)

func CheckClipboard(oldCount int) (string, error) {
	return "", errors.New("Not supported on Linux")
}

func GetClipboardCount() (int, error) {
	return int(0), errors.New("Not supported on Linux")
}
func GetFrontmostApp() (string, error) {
	return "", errors.New("Not supported on Linux")
}
func WaitForTime() {

}
