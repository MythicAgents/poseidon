package agentfunctions

import (
	"context"
	"path/filepath"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
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
			SupportedOS:      []string{},
			CommandIsBuiltin: true,
		},
		AssociatedBrowserScript: &agentstructs.BrowserScript{
			ScriptPath: filepath.Join(".", "poseidon", "browserscripts", "jobs.js"),
			Author:     "@its_a_feature_",
		},
		TaskFunctionParseArgString: func(ctx context.Context, args *agentstructs.PTTaskMessageArgsData, input string) error {
			return nil
		},
		TaskFunctionParseArgDictionary: func(ctx context.Context, args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return nil
		},
		TaskFunctionCreateTasking: func(ctx context.Context, task *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  task.Task.ID,
			}
			return response
		},
	})
}
