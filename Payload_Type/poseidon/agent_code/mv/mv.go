package mv

import (
	// Standard
	"encoding/json"
	"fmt"
	"os"
	"sync"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

var mu sync.Mutex

type Arguments struct {
	SourceFile      string `json:"source"`
	DestinationFile string `json:"destination"`
}

func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID

	var args Arguments
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

	if _, err = os.Stat(args.SourceFile); os.IsNotExist(err) {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"

		resp, _ := json.Marshal(msg)
		mu.Lock()
		profiles.TaskResponses = append(profiles.TaskResponses, resp)
		mu.Unlock()
		return
	}

	err = os.Rename(args.SourceFile, args.DestinationFile)

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
	msg.Completed = true
	msg.UserOutput = fmt.Sprintf("Moved %s to %s", args.SourceFile, args.DestinationFile)
	resp, _ := json.Marshal(msg)
	mu.Lock()
	profiles.TaskResponses = append(profiles.TaskResponses, resp)
	mu.Unlock()
	return
}
