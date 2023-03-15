package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "keys",
		HelpString:          "keys",
		Description:         "Interact with the linux keyring",
		Version:             1,
		MitreAttackMappings: []string{},
		Author:              "@xorrior",
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:          "command",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_CHOOSE_ONE,
				Description:   "Choose a way to interact with the keyring",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
					},
				},
				Choices: []string{"dumpsession", "dumpuser", "dumpprocess", "dumpthreads"},
			},
			{
				Name:          "keyword",
				Description:   "Name of the key to search for",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
						GroupName:           "search",
					},
				},
			},
			{
				Name:          "typename",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_CHOOSE_ONE,
				Description:   "Choose the type of the key",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     3,
						GroupName:           "search",
					},
				},
				Choices: []string{"keyring", "user", "login", "login", "session"},
			},
		},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{agentstructs.SUPPORTED_OS_LINUX},
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			if err := args.LoadArgsFromJSONString(input); err != nil {
				return err
			} else if groupName, err := args.GetParameterGroupName(); err != nil {
				return err
			} else if groupName == "search" {
				args.RemoveArg("command")
				args.AddArg(agentstructs.CommandParameter{
					Name:         "command",
					DefaultValue: "search",
					ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
						{
							GroupName: "search",
						},
					},
				})
				return nil
			} else {
				return nil
			}
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			if err := args.LoadArgsFromDictionary(input); err != nil {
				return err
			} else if groupName, err := args.GetParameterGroupName(); err != nil {
				return err
			} else if groupName == "search" {
				args.RemoveArg("command")
				args.AddArg(agentstructs.CommandParameter{
					Name:         "command",
					DefaultValue: "search",
					ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
						{
							GroupName: "search",
						},
					},
				})
				return nil
			} else {
				return nil
			}
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
