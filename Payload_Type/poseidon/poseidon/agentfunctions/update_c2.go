package agentfunctions

import (
	"errors"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "update_c2",
		Description:         "Update the C2 components within poseidon",
		HelpString:          "update_c2",
		Version:             1,
		Author:              "@its_a_feature_",
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "c2_name",
				ModalDisplayName: "C2 Profile Name",
				CLIName:          "c2",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
						GroupName:           "start/stop",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
						GroupName:           "update",
					},
				},
				Description: "The name of the c2 profile you want to configure",
			},
			{
				Name:             "action",
				CLIName:          "action",
				ModalDisplayName: "Action",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_CHOOSE_ONE,
				Choices:          []string{"start", "stop"},
				DefaultValue:     "start",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     2,
						GroupName:           "start/stop",
					},
				},
				Description: "Array of arguments to pass to the program.",
			},
			{
				Name:             "config_name",
				ModalDisplayName: "Config Name",
				CLIName:          "configName",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
						GroupName:           "update",
					},
				},
				Description: "The name of the c2 profile attribute you want to adjust",
			},
			{
				Name:             "config_value",
				ModalDisplayName: "New Config Value",
				CLIName:          "configValue",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     3,
						GroupName:           "update",
					},
				},
				Description: "The new value you want to use",
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			groupName, err := taskData.Args.GetParameterGroupName()
			if err != nil {

			}
			if groupName == "update" {
				taskData.Args.AddArg(agentstructs.CommandParameter{
					Name:             "action",
					ModalDisplayName: "action",
					CLIName:          "action",
					ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
					DefaultValue:     "update",
					ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
						{
							GroupName: "update",
						},
					},
				})
			}
			return response
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return args.LoadArgsFromDictionary(input)
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			if len(input) > 0 {
				return args.LoadArgsFromJSONString(input)
			} else {
				return errors.New("Must supply arguments")
			}
		},
	})
}
