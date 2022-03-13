package sleep

import (
	// Standard
	"encoding/json"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Args struct {
	Interval int `json:"interval"`
	Jitter   int `json:"jitter"`
}

//Run - interface method that retrieves a process list
func Run(task structs.Task) {

	args := Args{}
	err := json.Unmarshal([]byte(task.Params), &args)

	if err != nil {
		errResp := structs.Response{}
		errResp.Completed = true
		errResp.TaskID = task.TaskID
		errResp.Status = "error"
		errResp.UserOutput = err.Error()
		task.Job.SendResponses <- errResp
		return
	}
	output := ""
	if args.Interval >= 0 {
		output += task.Job.C2.SetSleepInterval(args.Interval)
	}
	if args.Jitter >= 0 && args.Jitter <= 100 {
		output += task.Job.C2.SetSleepJitter(args.Jitter)
	}
	resp := structs.Response{}
	resp.UserOutput = output
	resp.Completed = true
	resp.TaskID = task.TaskID
	task.Job.SendResponses <- resp
	return
}
