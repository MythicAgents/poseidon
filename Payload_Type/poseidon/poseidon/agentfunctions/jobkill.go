package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "jobkill",
		Description:         "Kill a job with the specified ID (from jobs command) - not all jobs are killable though.",
		HelpString:          "jobkill SOME-GUID-GOES-HERE",
		Version:             1,
		MitreAttackMappings: []string{"T1033"},
		SupportedUIFeatures: []string{"jobs:kill"},
		Author:              "@xorrior",
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
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
