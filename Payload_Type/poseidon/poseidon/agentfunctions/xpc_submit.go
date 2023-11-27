package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "xpc_submit",
		Description:         "Use xpc to submit a specific command and arguments for execution",
		HelpString:          "xpc_submit",
		Version:             1,
		MitreAttackMappings: []string{"T1559"},
		CommandParameters: []agentstructs.CommandParameter{

			{
				Name:             "program",
				ModalDisplayName: "Program to execute",
				CLIName:          "program",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Program/binary to execute",
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:           "submit",
						UIModalPosition:     1,
						ParameterIsRequired: true,
					},
					/*
						{
							GroupName:       "asuser",
							UIModalPosition: 2,
						},

					*/
				},
			},
			{
				Name:             "servicename",
				ModalDisplayName: "Service Name",
				CLIName:          "servicename",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Name of the service to create",
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:           "submit",
						UIModalPosition:     0,
						ParameterIsRequired: true,
					},
					/*
						{
							GroupName:       "asuser",
							UIModalPosition: 2,
						},

					*/
				},
			},
			/*
				{
					Name:             "uid",
					ModalDisplayName: "User UID",
					CLIName:          "uid",
					ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_NUMBER,
					Description:      "User UID to bootstrap for execution",
					DefaultValue:     0,
					ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
						{
							GroupName:       "asuser",
							UIModalPosition: 2,
						},
					},
				},

			*/
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
