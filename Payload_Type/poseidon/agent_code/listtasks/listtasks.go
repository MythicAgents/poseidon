package listtasks

import (
	// Standard
	"encoding/json"
	"sync"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

var mu sync.Mutex

type Listtasks interface {
	Result() map[string]interface{}
}

func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID

	r, err := getAvailableTasks()
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

	byteResult, err := json.MarshalIndent(r.Result(), "", "	")
	msg.UserOutput = string(byteResult)
	msg.Completed = true
	resp, _ := json.Marshal(msg)
	mu.Lock()
	profiles.TaskResponses = append(profiles.TaskResponses, resp)
	mu.Unlock()
	return
}
