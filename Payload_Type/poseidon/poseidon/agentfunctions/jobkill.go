package agentfunctions

import (
	"context"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "jobkill",
		Description:         "Kill a job with the specified ID (from jobs command) - not all jobs are killable though.",
		HelpString:          "jobkill SOME-GUID-GOES-HERE",
		Version:             1,
		MitreAttackMappings: []string{"T1033"},
		SupportedUIFeatures: []string{"jobs:kill", "task:job_kill"},
		Author:              "@xorrior",
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS:      []string{},
			CommandIsBuiltin: true,
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
