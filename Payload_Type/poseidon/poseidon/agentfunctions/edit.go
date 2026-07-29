package agentfunctions

import (
	"context"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "edit",
		Description:         "Open a UTF-8 text file up to 2 MB in Mythic's file editor.",
		HelpString:          "edit [file path]",
		Version:             1,
		Author:              "@its_a_feature_",
		SupportedUIFeatures: []string{"task_response:file_editor"},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS:        []string{agentstructs.SUPPORTED_OS_LINUX, agentstructs.SUPPORTED_OS_MACOS},
			CommandIsSuggested: true,
		},
		TaskFunctionCreateTasking: func(ctx context.Context, taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			return agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
		},
		TaskFunctionParseArgDictionary: func(ctx context.Context, args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return nil
		},
		TaskFunctionParseArgString: func(ctx context.Context, args *agentstructs.PTTaskMessageArgsData, input string) error {
			return nil
		},
	})
}
