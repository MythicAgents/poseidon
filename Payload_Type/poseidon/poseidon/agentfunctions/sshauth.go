package agentfunctions

import (
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
	"path/filepath"
	"strings"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name: "sshauth",
		Description: `SSH to specified host(s) using the designated credentials. 
You can also use this to execute a specific command on the remote hosts via SSH or use it to SCP files.`,
		HelpString:          "sshauth",
		Version:             1,
		Author:              "@xorrior",
		MitreAttackMappings: []string{"T1110.003"},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		AssociatedBrowserScript: &agentstructs.BrowserScript{
			ScriptPath: filepath.Join(".", "poseidon", "browserscripts", "sshauth_new.js"),
			Author:     "@djhohnstein",
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
						GroupName:           "scp-private-key",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
						GroupName:           "scp-plaintext-password",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
						GroupName:           "scp-private-key-credstore",
					},
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
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
						GroupName:           "run-command-private-key-credstore",
					},
				},
			},
			{
				Name:             "source",
				ModalDisplayName: "Source Filename",
				Description:      "If doing SCP, this is the source file",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     6,
						GroupName:           "scp-private-key",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     6,
						GroupName:           "scp-plaintext-password",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     6,
						GroupName:           "scp-private-key-credstore",
					},
				},
			},
			{
				Name:             "destination",
				ModalDisplayName: "Destination Filename",
				Description:      "If doing SCP, this is the destination file",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     7,
						GroupName:           "scp-private-key",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     7,
						GroupName:           "scp-plaintext-password",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     7,
						GroupName:           "scp-private-key-credstore",
					},
				},
			},
			{
				Name:             "private_key",
				ModalDisplayName: "Private Key",
				ParameterType:        agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
						GroupName:           "scp-private-key",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
						GroupName:           "run-command-private-key",
					},
				},
			},
			{
				Name:             "cred", // can't share `private_key` name
				ModalDisplayName: "Credential",
				ParameterType:        agentstructs.COMMAND_PARAMETER_TYPE_CHOOSE_ONE,
				Choices:              []string{""},
				DefaultValue:         "",
				DynamicQueryFunction: getCreds,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
						GroupName:           "scp-private-key-credstore",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
						GroupName:           "run-command-private-key-credstore",
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
						GroupName:           "scp-private-key",
					},
					{
						ParameterIsRequired: false,
						UIModalPosition:     5,
						GroupName:           "run-command-private-key",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     5,
						GroupName:           "scp-private-key-credstore",
					},
					{
						ParameterIsRequired: false,
						UIModalPosition:     5,
						GroupName:           "scp-plaintext-password",
					},
					{
						ParameterIsRequired: false,
						UIModalPosition:     5,
						GroupName:           "run-command-plaintext-password",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     5,
						GroupName:           "run-command-private-key-credstore",
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
						GroupName:           "scp-plaintext-password",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     3,
						GroupName:           "run-command-plaintext-password",
					},
				},
			},
			{
				Name:             "hosts",
				ModalDisplayName: "Array of CIDR notation for hosts",
				Description:      "Hosts that you will auth to",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_ARRAY,
				DefaultValue:     []string{},
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     4,
						GroupName:           "scp-plaintext-password",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     4,
						GroupName:           "scp-private-key-credstore",
					},
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
					{
						ParameterIsRequired: true,
						UIModalPosition:     4,
						GroupName:           "scp-private-key",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     4,
						GroupName:           "run-command-private-key-credstore",
					},
				},
			},
			{
				Name:             "command",
				ModalDisplayName: "Command to execute",
				Description:      "Command to execute on remote systems",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     6,
						GroupName:           "run-command-plaintext-password",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     6,
						GroupName:           "run-command-private-key",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     6,
						GroupName:           "run-command-private-key-credstore",
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
				if strings.Contains(groupName, "command") {
					displayParams += fmt.Sprintf(" to run a command")
				} else {
					displayParams += fmt.Sprintf(" to copy a file")
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

// getCreds dynamically fetches available credentials from Mythic
func getCreds(msg agentstructs.PTRPCDynamicQueryFunctionMessage) []string {
	rpcMessage := mythicrpc.MythicRPCCredentialSearchMessage{
		TaskID: msg.Callback,
	}

	rpcResponse, err := mythicrpc.SendMythicRPCCredentialSearch(rpcMessage)
	if err != nil || !rpcResponse.Success {
		return []string{"Error fetching credentials"}
	}

	choices := []string{}
	for _, cred := range rpcResponse.Credentials {
		if cred.Account != nil && cred.Realm != nil {
			choices = append(choices, fmt.Sprintf("Account: %s, Realm: %s, Credential: %s", *cred.Account, *cred.Realm, *cred.Credential))
		}
	}

	return choices
}
