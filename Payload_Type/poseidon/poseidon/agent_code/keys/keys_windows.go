//go:build windows
// +build windows

package keys

import "errors"

// KeyContents - struct that represent raw key contents
type WindowsKeyInformation struct {
	KeyType string
	KeyData []byte
}

// Type - The type of key information. Keyring or keychain
func (l *WindowsKeyInformation) Type() string {
	return l.KeyType
}

// KeyData - Retrieve the keydata as a raw json string
func (l *WindowsKeyInformation) Data() []byte {
	return l.KeyData
}

func getkeydata(opts Options) (WindowsKeyInformation, error) {
	//Check if the types are available
	d := WindowsKeyInformation{}
	return d, errors.New("Not implemented")
}
