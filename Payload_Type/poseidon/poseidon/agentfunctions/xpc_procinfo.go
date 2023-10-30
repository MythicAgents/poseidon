package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "xpc_procinfo",
		Description:         "Use xpc to get the process information for a specific pid",
		HelpString:          "xpc_procinfo",
		Version:             1,
		MitreAttackMappings: []string{"T1559"},
		CommandParameters: []agentstructs.CommandParameter{
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
