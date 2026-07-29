package agentfunctions

import (
	"context"
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
			SupportedOS:        []string{agentstructs.SUPPORTED_OS_MACOS},
			CommandIsSuggested: true,
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
						GroupName:           "Default",
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
						GroupName:           "Default",
					},
				},
			},
			{
				Name:                   "credential",
				CLIName:                "credential",
				ModalDisplayName:       "Credential",
				Description:            "Credential store username/password.",
				ParameterType:          agentstructs.COMMAND_PARAMETER_TYPE_CREDENTIAL,
				DefaultValue:           "",
				LimitCredentialsByType: []string{"plaintext"},
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
						GroupName:           "Stored Credential",
					},
				},
			},
		},
		TaskFunctionCreateTasking: func(ctx context.Context, taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			groupName, err := taskData.Args.GetParameterGroupName()
			if err != nil {
				response.Error = err.Error()
				response.Success = false
				return response
			}
			if groupName == "Default" {
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
				displayString := fmt.Sprintf("for %s with password \"%s\"",
					taskData.Task.RevertKeywords(userString, "username"),
					taskData.Task.RevertKeywords(passwordString, "password"))
				response.DisplayParams = &displayString
			} else {
				credential, err := taskData.Args.GetCredentialArg("credential")
				if err != nil {
					response.Error = err.Error()
					response.Success = false
					return response
				}
				displayString := fmt.Sprintf("-credential %s", taskData.Task.RevertKeywords(credential, "credential"))
				response.DisplayParams = &displayString
				taskData.Args.RemoveArg("credential")
				taskData.Args.RemoveArg("username")
				taskData.Args.RemoveArg("password")
				taskData.Args.AddArg(agentstructs.CommandParameter{
					Name:         "username",
					DefaultValue: credential.Account,
					ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
						{
							ParameterIsRequired: false,
							UIModalPosition:     1,
							GroupName:           "Stored Credential",
						},
					},
				})
				taskData.Args.AddArg(agentstructs.CommandParameter{
					Name:         "password",
					DefaultValue: credential.Credential,
					ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
						{
							ParameterIsRequired: false,
							UIModalPosition:     2,
							GroupName:           "Stored Credential",
						},
					},
				})
			}
			return response
		},
		TaskFunctionParseArgDictionary: func(ctx context.Context, args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return args.LoadArgsFromDictionary(input)
		},
		TaskFunctionParseArgString: func(ctx context.Context, args *agentstructs.PTTaskMessageArgsData, input string) error {
			return nil
		},
	})
}
