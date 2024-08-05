package agentfunctions

import (
	"errors"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "sudo",
		Description:         "Attempt to execute a command in a root context with a supplied username/password. If that's not known, prompt text and a prompt icon path can be used to cause a popup for the user",
		HelpString:          "sudo -username bob -password superSecretPa55w0rd -command /usr/bin/id",
		Version:             1,
		Author:              "@its_a_feature_",
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{agentstructs.SUPPORTED_OS_MACOS},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "username",
				ModalDisplayName: "Username",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     1,
					},
				},
				Description: "Username to authenticate as",
			},
			{
				Name:             "password",
				ModalDisplayName: "Password",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     2,
					},
				},
				Description: "Password for the specified Username",
			},
			{
				Name:             "command",
				ModalDisplayName: "Command",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     3,
					},
				},
				Description: "Command to execute with privileges",
			},
			{
				Name:             "args",
				ModalDisplayName: "Args",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_ARRAY,
				Description:      "Any args you want to pass to the program specified by command",
				DefaultValue:     []string{},
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     4,
					},
				},
			},
			{
				Name:             "prompt_text",
				ModalDisplayName: "Prompt Text",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     5,
					},
				},
				Description: "Text to display to the user when prompting to execute as root",
			},
			{
				Name:             "prompt_icon_path",
				ModalDisplayName: "Prompt Icon Path",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     6,
					},
				},
				Description: "Path to the icon to use as part of a popup dialog asking the user to authenticate",
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
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
