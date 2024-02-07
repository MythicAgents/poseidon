package agentfunctions

import (
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "unlink_webshell",
		Description:         "Unlink a webshell connection.",
		HelpString:          "unlink_webshell",
		Version:             1,
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{},
		Author:              "@its_a_feature_",
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:          "connection",
				Description:   "Connection info for unlinking",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_LINK_INFO,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:           "Modal Selection",
						ParameterIsRequired: true,
					},
				},
			},
			{
				Name:          "connectionUUID",
				Description:   "Existing UUID within Poseidon to unlink",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName:           "Explicit UUID",
						ParameterIsRequired: true,
					},
				},
			},
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			args.SetArgValue("connectionUUID", input)
			return nil
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return args.LoadArgsFromDictionary(input)
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			groupName, err := taskData.Args.GetParameterGroupName()
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			if groupName == "Explicit UUID" {
				connectionString, err := taskData.Args.GetStringArg("connectionUUID")
				if err != nil {
					response.Success = false
					response.Error = err.Error()
					return response
				} else {
					taskData.Args.RemoveArg("connectionUUID")
					taskData.Args.RemoveArg("connection")
					taskData.Args.AddArg(agentstructs.CommandParameter{
						Name:          "connection",
						DefaultValue:  connectionString,
						ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
						ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
							{
								GroupName: "Explicit UUID",
							},
						},
					})
					displayString := fmt.Sprintf("from %s", connectionString)
					response.DisplayParams = &displayString
				}
			} else {
				if connectionInfo, err := taskData.Args.GetLinkInfoArg("connection"); err != nil {
					response.Success = false
					response.Error = err.Error()
				} else if connectionInfo.CallbackUUID == "" {
					response.Success = false
					response.Error = "Failed to find callback UUID in connection information"
				} else {
					taskData.Args.RemoveArg("connection")
					taskData.Args.AddArg(agentstructs.CommandParameter{
						Name:          "connection",
						DefaultValue:  connectionInfo.CallbackUUID,
						ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
						ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
							{
								GroupName: "Modal Selection",
							},
						},
					})
					displayString := fmt.Sprintf("from %s", connectionInfo.CallbackUUID)
					response.DisplayParams = &displayString
				}
			}

			return response
		},
	})
}
