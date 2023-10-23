package agentfunctions

import (
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                  "test_password",
		Description:           "Use OpenDirectory API to test a user's password.",
		HelpString:            "test_password -username username -password password",
		Version:               1,
		Author:                "@its_a_feature",
		MitreAttackMappings:   []string{},
		SupportedUIFeatures:   []string{},
		NeedsAdminPermissions: true,
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{agentstructs.SUPPORTED_OS_MACOS},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "username",
				CLIName:          "username",
				ModalDisplayName: "Username",
				Description:      "Username of the user to test the password for.",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
					},
				},
			},
			{
				Name:             "password",
				CLIName:          "password",
				ModalDisplayName: "Password",
				Description:      "Password for the user to test against.",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
					},
				},
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			userString, err := taskData.Args.GetStringArg("username")
			if err != nil {
				response.Error = err.Error()
				response.Success = false
				return response
			}
			passwordString, err := taskData.Args.GetStringArg("password")
			if err != nil {
				response.Error = err.Error()
				response.Success = false
				return response
			}
			displayString := fmt.Sprintf("for %s with password \"%s\"", userString, passwordString)
			response.DisplayParams = &displayString
			return response
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return args.LoadArgsFromDictionary(input)
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			return nil
		},
	})
}
