// +build windows
package jxa

import (
	"errors"
)

type JxaRunWindows struct {
	Successful bool
	Resultstring string
}

func (j *JxaRunWindows) Success() bool {
	return j.Successful
}

func (j *JxaRunWindows) Result() string {
	return j.Resultstring
}


func runCommand(encpayload string) (JxaRunWindows, error) {
	n := JxaRunWindows{}
	n.Resultstring = ""
	n.Successful = false
	return n, errors.New("Not implemented")
}
