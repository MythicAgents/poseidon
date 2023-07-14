package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "xpc",
		Description:         "Use xpc to execute routines with launchd or communicate with another service/process.",
		HelpString:          "xpc",
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
				},
			},
			{
				Name:             "load",
				ModalDisplayName: "Flag to indicate the load command",
				CLIName:          "load",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				Description:      "Must be True to run 'load' with a file path to a plist file",
				DefaultValue:     true,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:       "load",
						UIModalPosition: 1,
					},
				},
			},
			{
				Name:             "unload",
				ModalDisplayName: "Flag to indicate the unload command",
				CLIName:          "unload",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				Description:      "Must be True to run 'unload' with a file path to a plist file",
				DefaultValue:     true,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:       "unload",
						UIModalPosition: 1,
					},
				},
			},
			{
				Name:             "file",
				ModalDisplayName: "Path to the file to load/unload on target",
				CLIName:          "file",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Path to the plist file on disk to load/unload",
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:           "load",
						UIModalPosition:     0,
						ParameterIsRequired: true,
					},
					{
						GroupName:           "unload",
						UIModalPosition:     0,
						ParameterIsRequired: true,
					},
				},
			},
			{
				Name:             "servicename",
				ModalDisplayName: "Service Name",
				CLIName:          "servicename",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Name of the service to communicate with. Used with the submit, send, start/stop commands",
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:           "send",
						UIModalPosition:     0,
						ParameterIsRequired: true,
					},
					{
						GroupName:           "submit",
						UIModalPosition:     0,
						ParameterIsRequired: true,
					},
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
						GroupName:           "status",
						UIModalPosition:     0,
						ParameterIsRequired: true,
					},
					{
						GroupName:           "list",
						UIModalPosition:     0,
						ParameterIsRequired: false,
					},
				},
			},
			{
				Name:             "pid",
				ModalDisplayName: "PID",
				CLIName:          "pid",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_NUMBER,
				Description:      "PID of the process to target",
				DefaultValue:     0,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:       "procinfo",
						UIModalPosition: 1,
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
			return response
		},
	})
}
