package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"path/filepath"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "ps",
		Description:         "Get a process listing (with optional regex filtering).",
		HelpString:          "ps [regex name mathcing]",
		Version:             1,
		Author:              "@djhohnstein, @xorroir, @its_a_feature_",
		MitreAttackMappings: []string{"T1057"},
		SupportedUIFeatures: []string{"process_browser:list"},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		AssociatedBrowserScript: &agentstructs.BrowserScript{
			ScriptPath: filepath.Join(".", "poseidon", "browserscripts", "ps_new.js"),
			Author:     "@djhohnstein",
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "regex_filter",
				CLIName:          "regex_filter",
				ModalDisplayName: "Regex Filter",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Regular expression filter to limit which processes are returned",
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
					},
				},
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
			return nil
		},
	})
}
