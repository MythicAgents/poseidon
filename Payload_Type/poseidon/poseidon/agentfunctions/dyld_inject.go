package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "dyld_inject",
		HelpString:          "dyld_inject",
		Description:         "Spawn an application/binary and inject a dylib into application with the DYLD_INSERT_LIBRARIES environment variable",
		Version:             1,
		MitreAttackMappings: []string{"T1574.006"},
		Author:              "@xorrior @_r3ggi",
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:          "application",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:   "Path to the target application/binary",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
					},
				},
			},
			{
				Name:             "dylibpath",
				ModalDisplayName: "Dylib Path",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Path to the dylib on disk that will be injected into the target application",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
					},
				},
			},
			{
				Name:          "hideApp",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				DefaultValue:  false,
				Description:   "If true, launch the application with the kLSLaunchAndHide flag set. If false, use the kLSLaunchDefaults flag",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     3,
					},
				},
			},
		},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{agentstructs.SUPPORTED_OS_MACOS},
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			return args.LoadArgsFromJSONString(input)
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return args.LoadArgsFromDictionary(input)
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
