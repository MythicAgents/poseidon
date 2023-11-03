//go:build linux

package sudo

import "github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"

func sudoWithPromptOption(task structs.Task, args Arguments) {
	msg := task.NewResponse()
	msg.SetError("Not Implemented")
	task.Job.SendResponses <- msg
	return
}
