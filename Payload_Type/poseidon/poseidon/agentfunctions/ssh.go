package agentfunctions

import (
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"strings"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "ssh",
		Description:         `SSH to host using the designated credentials and open a PTY without spawning ssh`,
		HelpString:          "ssh",
		Version:             1,
		Author:              "@its_a_feature_",
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{"task_response:interactive"},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "username",
				ModalDisplayName: "Username",
				Description:      "Authenticate to the designated hosts using this username",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
						GroupName:           "run-command-plaintext-password",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
						GroupName:           "run-command-private-key",
					},
				},
			},
			{
				Name:             "private_key",
				ModalDisplayName: "Path to Private key on disk",
				Description:      "Authenticate to the designated hosts using this private key",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
						GroupName:           "run-command-private-key",
					},
				},
			},
			{
				Name:             "port",
				ModalDisplayName: "SSH Port",
				Description:      "SSH Port if different than 22",
				DefaultValue:     22,
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_NUMBER,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     5,
						GroupName:           "run-command-private-key",
					},
					{
						ParameterIsRequired: false,
						UIModalPosition:     5,
						GroupName:           "run-command-plaintext-password",
					},
				},
			},
			{
				Name:             "password",
				ModalDisplayName: "Plaintext Password",
				Description:      "Authenticate to the designated hosts using this password",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     3,
						GroupName:           "run-command-plaintext-password",
					},
				},
			},
			{
				Name:             "host",
				ModalDisplayName: "Hostname or IP",
				Description:      "Host that you will auth to",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:     "127.0.0.1",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     4,
						GroupName:           "run-command-plaintext-password",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     4,
						GroupName:           "run-command-private-key",
					},
				},
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			displayParams := ""
			if username, err := taskData.Args.GetStringArg("username"); err != nil {
				response.Success = false
				response.Error = err.Error()
			} else if groupName, err := taskData.Args.GetParameterGroupName(); err != nil {
				response.Success = false
				response.Error = err.Error()
			} else {
				displayParams += fmt.Sprintf("as %s ", username)
				if strings.Contains(groupName, "private-key") {
					// authing with private key
					displayParams += fmt.Sprintf("with a private key")
				} else {
					// authing with plaintext password
					displayParams += fmt.Sprintf("with a plaintext password")
				}
				response.DisplayParams = &displayParams
			}
			return response
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return args.LoadArgsFromDictionary(input)
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			return args.LoadArgsFromJSONString(input)
		},
	})
}
