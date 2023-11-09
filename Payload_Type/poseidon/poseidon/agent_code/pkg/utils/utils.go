package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)
import "log"

var (
	// debug is used for deciding to print debug messages or not
	debug bool
	// debugString
	debugString string
	// SeededRand is used when generating a random value for EKE
	SeededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func init() {
	if debugString == "" || strings.ToLower(debugString) == "false" {
		debug = false
	} else {
		debug = true
		fmt.Printf("debug string: %s\n", debugString)
	}
}

func PrintDebug(msg string) {
	if debug {
		log.Print(msg)
	}
}

func GenerateSessionID() string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 20)
	for i := range b {
		b[i] = letterBytes[SeededRand.Intn(len(letterBytes))]
	}
	return string(b)
}

func RandomNumInRange(limit int) int {
	return SeededRand.Intn(limit)
}
