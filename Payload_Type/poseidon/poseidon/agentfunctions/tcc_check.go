package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                  "tcc_check",
		Description:           "Use MDQuery APIs to check for various TCC permissions.",
		HelpString:            "tcc_check",
		Version:               1,
		Author:                "@its_a_feature, @slyd0g",
		MitreAttackMappings:   []string{},
		SupportedUIFeatures:   []string{},
		NeedsAdminPermissions: true,
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{agentstructs.SUPPORTED_OS_MACOS},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "user",
				CLIName:          "user",
				ModalDisplayName: "User to check access against",
				Description:      "If no user is supplied, current user context is checked.",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
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
			userString, err := taskData.Args.GetStringArg("user")
			if err != nil {
				response.Error = err.Error()
				response.Success = false
			} else {
				response.DisplayParams = &userString
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
