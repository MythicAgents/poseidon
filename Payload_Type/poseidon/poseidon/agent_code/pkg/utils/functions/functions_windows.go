//go:build windows
package functions


import (
	"fmt"
	"os"
	"runtime"
	"golang.org/x/sys/windows"
)

func isElevated() bool {
	return true
}
func getArchitecture() string {
	return runtime.GOARCH
}
func getProcessName() string {
	name, err := os.Executable()
	if err != nil {
		return ""
	} else {
		return name
	}
}
func getDomain() string {
	// TODO: implement me
	return ""
}
func getStringFromBytes(data [65]byte) string {
	stringData := make([]byte, 0, 0)
	for i := range data {
		if data[i] == 0 {
			return string(stringData[:])
		} else {
			stringData = append(stringData, data[i])
		}
	}
	return string(stringData[:])
}
func getOS() string {
	verInfo := windows.RtlGetVersion()
	return fmt.Sprintf("%d.%d (Build %d)", verInfo.MajorVersion, verInfo.MinorVersion, verInfo.BuildNumber) 
}

func getUser() string {
	name := make([]uint16, 128)
	nameSize := uint32(len(name))
	err := windows.GetUserNameEx(windows.NameSamCompatible, &name[0], &nameSize)
	if err != nil {
		return ""
	}
	return windows.UTF16ToString(name)
}

func getPID() int {
	return os.Getpid()
}
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	} else {
		return hostname
	}
}
