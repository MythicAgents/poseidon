//go:build linux
// +build linux

package clipboard_monitor

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

func CheckClipboard(oldCount int) (string, error) {
	return "", nil
}

func GetClipboardCount() (int, error) {
	return int(0)
}
func GetFrontmostApp() (string, error) {
	return "", nil
}
func WaitForTime() {

}
