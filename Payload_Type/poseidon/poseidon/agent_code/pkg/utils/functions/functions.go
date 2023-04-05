package functions

import (
	"net"
	"sort"
)

// GetCurrentIPAddress - the current IP address of the system
func GetCurrentIPAddress() []string {
	if addrs, err := net.InterfaceAddrs(); err != nil {
		return []string{"127.0.0.1"}
	} else {
		ipAddresses := []string{}
		for _, address := range addrs {
			addrString := address.(*net.IPNet).IP
			if !addrString.IsLoopback() {
				ipAddresses = append(ipAddresses, addrString.String())
			}
		}
		sort.Sort(sort.StringSlice(ipAddresses))
		return ipAddresses
	}
}

func SliceContains[V string | int](source []V, check V) bool {
	for _, v := range source {
		if check == v {
			return true
		}
	}
	return false
}

func IsElevated() bool {
	return isElevated()
}
func GetArchitecture() string {
	return getArchitecture()
}
func GetDomain() string {
	return getDomain()
}
func GetOS() string {
	return getOS()
}
func GetProcessName() string {
	return getProcessName()
}
func GetUser() string {
	return getUser()
}
func GetPID() int {
	return getPID()
}
func GetHostname() string {
	return getHostname()
}
