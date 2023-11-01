package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"path/filepath"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "clipboard",
		Description:         "Get the contents of the clipboard",
		HelpString:          "clipboard -duration -1",
		Version:             1,
		Author:              "@its_a_feature_",
		MitreAttackMappings: []string{"T1115"},
		SupportedUIFeatures: []string{"clipboard:list"},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{agentstructs.SUPPORTED_OS_MACOS},
		},
		AssociatedBrowserScript: &agentstructs.BrowserScript{
			ScriptPath: filepath.Join(".", "poseidon", "browserscripts", "clipboard.js"),
			Author:     "@its_a_feature_",
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "read",
				ModalDisplayName: "Read Types",
				DefaultValue:     []string{"public.utf8-plain-text"},
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_ARRAY,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     1,
					},
				},
				Description: "The various types to fetch from the clipboard. Using * will fetch the content of everything on the clipboard (this could be a lot)",
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			return response
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return args.LoadArgsFromDictionary(input)
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			return args.LoadArgsFromJSONString(input)
		},
	})
}
