package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "xpc_load",
		Description:         "Use xpc to load a new launch agent or launch daemon",
		HelpString:          "xpc_load",
		Version:             1,
		MitreAttackMappings: []string{"T1559"},
		CommandParameters: []agentstructs.CommandParameter{

			{
				Name:             "file",
				ModalDisplayName: "Path to the file to load on target",
				CLIName:          "file",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Path to the plist file on disk to load",
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:           "load",
						UIModalPosition:     0,
						ParameterIsRequired: true,
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
