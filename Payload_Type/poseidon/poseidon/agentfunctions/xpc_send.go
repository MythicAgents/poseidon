package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "xpc_send",
		Description:         "Use xpc to send data to an xpc service",
		HelpString:          "xpc_send",
		Version:             1,
		MitreAttackMappings: []string{"T1559"},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "servicename",
				ModalDisplayName: "Service Name",
				CLIName:          "servicename",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Name of the service to communicate with",
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:           "send",
						UIModalPosition:     0,
						ParameterIsRequired: true,
					},
				},
			},
			{
				Name:             "data",
				ModalDisplayName: "Data to send",
				CLIName:          "data",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "base64 encoded JSON of data to send to a target service",
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:       "send",
						UIModalPosition: 1,
					},
				},
			},
		},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{agentstructs.SUPPORTED_OS_MACOS},
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			err := args.LoadArgsFromJSONString(input)
			if err != nil {
				return err
			}
			groupName, err := args.GetParameterGroupName()
			if err != nil {
				return err
			}
			args.AddArg(agentstructs.CommandParameter{
				Name:          "command",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:  groupName,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName: groupName,
					},
				},
			})
			return nil
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			err := args.LoadArgsFromDictionary(input)
			if err != nil {
				return err
			}
			groupName, err := args.GetParameterGroupName()
			if err != nil {
				return err
			}
			args.AddArg(agentstructs.CommandParameter{
				Name:          "command",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:  groupName,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName: groupName,
					},
				},
			})
			return nil
		},
		TaskFunctionCreateTasking: func(task *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  task.Task.ID,
			}
			commandName := "xpc"
			response.CommandName = &commandName
			return response
		},
	})
}
