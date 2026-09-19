//go:build cgo

package main

import (
	"C"
)

//export RunMain
func RunMain() {
	main()
}
