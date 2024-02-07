package agentfunctions

import (
	"errors"
	"fmt"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/mythicrpc"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "link_webshell",
		Description:         "Link to an agent using the webshell p2p profile",
		HelpString:          "link_webshell",
		Version:             1,
		Author:              "@its_a_feature_",
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:          "connection",
				CLIName:       "connectionDictionary",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_CONNECTION_INFO,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
					},
				},
				Description: "Mythic's detailed connection information",
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			connectionInfo, err := taskData.Args.GetConnectionInfoArg("connection")
			if err != nil {
				logging.LogError(err, "Failed to get connection information")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			err = taskData.Args.RemoveArg("connection")
			if err != nil {
				logging.LogError(err, "Failed to remove connection data")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			/*
				{
					"host":"SPOOKY.LOCAL",
					"c2_profile":{
						"name":"poseidon_tcp",
						"parameters":{
							"AESPSK":{"crypto_type":"aes256_hmac","enc_key":"7JSBbGON1cHI4xtpxR0M41qQulCBD+DgyABLr6hpjFc=","dec_key":"7JSBbGON1cHI4xtpxR0M41qQulCBD+DgyABLr6hpjFc="},
							"port":"8085",
							"killdate":"2024-09-06",
							"encrypted_exchange_check":"true"
						}
					},
					"callback_uuid":"150732f5-65c6-424c-a97d-e2d2888f7856",
					"agent_uuid":"80844d19-9bfc-47f9-b9af-c6b9144c0fdc", // this or callback_uuid, not both
				}
			*/
			taskData.Args.AddArg(agentstructs.CommandParameter{
				Name:          "url",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:  connectionInfo.C2ProfileInfo.Parameters["url"],
			})
			taskData.Args.AddArg(agentstructs.CommandParameter{
				Name:          "user_agent",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:  connectionInfo.C2ProfileInfo.Parameters["user_agent"],
			})
			taskData.Args.AddArg(agentstructs.CommandParameter{
				Name:          "query_param",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:  connectionInfo.C2ProfileInfo.Parameters["query_param"],
			})
			taskData.Args.AddArg(agentstructs.CommandParameter{
				Name:          "cookie_name",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:  connectionInfo.C2ProfileInfo.Parameters["cookie_name"],
			})
			if connectionInfo.AgentUUID != "" {
				response.Success = false
				response.Error = "Must connect to an existing Callback. If one isn't available, use the Payload's page to create a new callback based on the payload you used."
				return response
			} else {
				// we have the callback uuid and need the payload uuid
				callbackSearchResponse, err := mythicrpc.SendMythicRPCCallbackSearch(mythicrpc.MythicRPCCallbackSearchMessage{
					AgentCallbackID:    taskData.Callback.ID,
					SearchCallbackUUID: &connectionInfo.CallbackUUID,
				})
				if err != nil {
					logging.LogError(err, "Failed to search callbacks data")
					response.Success = false
					response.Error = err.Error()
					return response
				}
				if !callbackSearchResponse.Success {
					logging.LogError(err, "Failed to search callbacks data")
					response.Success = false
					response.Error = callbackSearchResponse.Error
					return response
				}
				if len(callbackSearchResponse.Results) == 0 {
					logging.LogError(err, "Failed to remove connection data")
					response.Success = false
					response.Error = err.Error()
					return response
				}
				taskData.Args.AddArg(agentstructs.CommandParameter{
					Name:          "cookie_value",
					ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
					DefaultValue:  callbackSearchResponse.Results[0].RegisteredPayloadUUID,
				})
				taskData.Args.AddArg(agentstructs.CommandParameter{
					Name:          "target_uuid",
					ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
					DefaultValue:  connectionInfo.CallbackUUID,
				})
			}
			displayString := fmt.Sprintf("%s", connectionInfo.C2ProfileInfo.Parameters["url"].(string))
			response.DisplayParams = &displayString

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
