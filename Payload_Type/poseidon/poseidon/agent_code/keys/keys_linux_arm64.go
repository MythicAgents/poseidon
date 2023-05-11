//go:build linux && arm64
// +build linux,arm64

package keys

import "errors"

// KeyContents - struct that represent raw key contents
type LinuxKeyInformation struct {
	KeyType string
	KeyData []byte
}

// Type - The type of key information. Keyring or keychain
func (l *LinuxKeyInformation) Type() string {
	return l.KeyType
}

// KeyData - Retrieve the keydata as a raw json string
func (l *LinuxKeyInformation) Data() []byte {
	return l.KeyData
}

func getkeydata(opts Options) (LinuxKeyInformation, error) {
	//Check if the types are available
	d := LinuxKeyInformation{}
	return d, errors.New("Not implemented for ARM")
}
