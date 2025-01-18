//go:build windows
// +build windows

package clipboard_monitor

import (
	"errors"
)

func CheckClipboard(oldCount int) (string, error) {
	return "", errors.New("Not supported on Windows")
}

func GetClipboardCount() (int, error) {
	return int(0), errors.New("Not supported on Windows")
}
func GetFrontmostApp() (string, error) {
	return "", errors.New("Not supported on Windows")
}
func WaitForTime() {

}
