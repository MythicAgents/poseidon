package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "xpc_service",
		Description:         "Use xpc to manipulate or list existing services",
		HelpString:          "xpc_service",
		Version:             1,
		MitreAttackMappings: []string{"T1559"},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "list",
				ModalDisplayName: "List",
				CLIName:          "list",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				Description:      "Flag to indicate asking launchd to list running services",
				DefaultValue:     true,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:       "list",
						UIModalPosition: 0,
					},
				},
			},
			{
				Name:             "start",
				ModalDisplayName: "Start",
				CLIName:          "start",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				Description:      "Flag to indicate asking launchd to start a service",
				DefaultValue:     true,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:       "start",
						UIModalPosition: 1,
					},
				},
			},
			{
				Name:             "stop",
				ModalDisplayName: "Stop",
				CLIName:          "stop",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				Description:      "Flag to indicate asking launchd to stop a service",
				DefaultValue:     true,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:       "stop",
						UIModalPosition: 1,
					},
				},
			},
			{
				Name:             "enable",
				ModalDisplayName: "Enable",
				CLIName:          "enable",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				Description:      "Flag to indicate asking launchd to enable a service",
				DefaultValue:     true,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:       "enable",
						UIModalPosition: 1,
					},
				},
			},
			{
				Name:             "disable",
				ModalDisplayName: "Disable",
				CLIName:          "disable",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				Description:      "Flag to indicate asking launchd to disable a service",
				DefaultValue:     true,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:       "disable",
						UIModalPosition: 1,
					},
				},
			},
			{
				Name:             "remove",
				ModalDisplayName: "Remove",
				CLIName:          "remove",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				Description:      "Flag to indicate asking launchd to remove the specified service",
				DefaultValue:     true,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:       "remove",
						UIModalPosition: 0,
					},
				},
			},
			{
				Name:             "print",
				ModalDisplayName: "Print",
				CLIName:          "print",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				Description:      "Flag to indicate asking launchd to print information about the specified service or all services",
				DefaultValue:     true,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:       "print",
						UIModalPosition: 0,
					},
				},
			},
			{
				Name:             "dumpstate",
				ModalDisplayName: "DumpState",
				CLIName:          "dumpstate",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				Description:      "Flag to indicate asking launchd to print information about the specified service or all services",
				DefaultValue:     true,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:       "dumpstate",
						UIModalPosition: 0,
					},
				},
			},
			{
				Name:             "servicename",
				ModalDisplayName: "Service Name",
				CLIName:          "servicename",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Name of the service to communicate with. Used with the submit, send, start/stop, print commands",
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:           "start",
						UIModalPosition:     0,
						ParameterIsRequired: true,
					},
					{
						GroupName:           "stop",
						UIModalPosition:     0,
						ParameterIsRequired: true,
					},
					{
						GroupName:           "enable",
						UIModalPosition:     0,
						ParameterIsRequired: true,
					},
					{
						GroupName:           "disable",
						UIModalPosition:     0,
						ParameterIsRequired: true,
					},
					{
						GroupName:           "list",
						UIModalPosition:     0,
						ParameterIsRequired: false,
					},
					{
						GroupName:           "remove",
						UIModalPosition:     0,
						ParameterIsRequired: true,
					},
					{
						GroupName:           "print",
						UIModalPosition:     0,
						ParameterIsRequired: false,
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
