package functions

import (
	"net"
)

//GetCurrentIPAddress - the current IP address of the system
func GetCurrentIPAddress() string {
	addrs, err := net.InterfaceAddrs()
	currIP := "127.0.0.1"
	if err != nil {
		return currIP
	}
	for _, address := range addrs {

		// return the first IPv4 address that's not localhost
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				//fmt.Println("Current IP address : ", ipnet.IP.String())
				if ipnet.IP.String() == currIP {
					continue
				}
				return ipnet.IP.String()
			}
		}
	}

	return currIP
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
