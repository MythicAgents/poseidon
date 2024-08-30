//go:build linux

package functions

/*
#include <unistd.h>
int UpdateEUID();
int UpdateEUID(){
    uid_t euid = geteuid();
    uid_t uid = getuid();
    if(euid != uid){
        setuid(euid);
    }
    gid_t egid = getegid();
    gid_t gid = getgid();
    if(egid != gid){
        setgid(egid);
    }
	uid_t finalUID = getuid();
    return finalUID;
}
*/
import "C"
import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"unicode/utf16"

	"golang.org/x/sys/unix"
)

func isElevated() bool {
	uid := C.UpdateEUID()
	return uid == 0
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
	u := unix.Utsname{}
	err := unix.Uname(&u)
	if err != nil {
		return fmt.Sprintf("Unknown Error: %v", err)
	}
	return getStringFromBytes(u.Sysname) + "\n" + getStringFromBytes(u.Nodename) + "\n" + getStringFromBytes(u.Release) + "\n" + getStringFromBytes(u.Version) + "\n" + getStringFromBytes(u.Machine)
}
func getUser() string {
	uid := C.UpdateEUID()
	currentUser, err := user.LookupId(strconv.Itoa(int(uid)))
	//currentUser, err := user.Current()
	if err != nil {
		return ""
	} else {
		return currentUser.Username
	}
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

// Helper function to convert DWORD byte counts to
// human readable sizes.
func UINT32ByteCountDecimal(b uint32) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint32(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float32(b)/float32(div), "kMGTPE"[exp])
}

// Helper function to convert LARGE_INTEGER byte
//
//	counts to human readable sizes.
func UINT64ByteCountDecimal(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

// Helper function to build a string from a WCHAR string
func UTF16ToString(s []uint16) []string {
	var results []string
	cut := 0
	for i, v := range s {
		if v == 0 {
			if i-cut > 0 {
				results = append(results, string(utf16.Decode(s[cut:i])))
			}
			cut = i + 1
		}
	}
	if cut < len(s) {
		results = append(results, string(utf16.Decode(s[cut:])))
	}
	return results
}
