package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"path/filepath"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "jobs",
		Description:         "List running/killable jobs",
		HelpString:          "jobs",
		Version:             1,
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{},
		Author:              "@xorrior",
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		AssociatedBrowserScript: &agentstructs.BrowserScript{
			ScriptPath: filepath.Join(".", "poseidon", "browserscripts", "jobs.js"),
			Author:     "@its_a_feature_",
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			return nil
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return nil
		},
		TaskFunctionCreateTasking: func(task *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  task.Task.ID,
			}
			return response
		},
	})
}
