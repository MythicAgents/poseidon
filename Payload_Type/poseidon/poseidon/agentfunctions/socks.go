package agentfunctions

import (
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
	"github.com/MythicMeta/MythicContainer/rabbitmq"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "socks",
		Description:         "Start or Stop SOCKS5.",
		HelpString:          "socks",
		Version:             1,
		Author:              "@xorrior",
		MitreAttackMappings: []string{"T1572"},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "action",
				ModalDisplayName: "Action",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_CHOOSE_ONE,
				Choices:          []string{"start", "stop", "flush"},
				DefaultValue:     "start",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
					},
				},
				Description: "Start or Stop socks through this callback",
			},
			{
				Name:             "port",
				ModalDisplayName: "Local Mythic Port for SOCKS5",
				DefaultValue:     7000,
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_NUMBER,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
					},
				},
				Description: "Port number on Mythic server to open for SOCKS5",
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			if action, err := taskData.Args.GetStringArg("action"); err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			} else if port, err := taskData.Args.GetNumberArg("port"); err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			} else {
				displayString := fmt.Sprintf("%s on port %.0f", action, port)
				response.DisplayParams = &displayString
				if action == "start" {
					if socksResponse, err := mythicrpc.SendMythicRPCProxyStart(mythicrpc.MythicRPCProxyStartMessage{
						PortType:  rabbitmq.CALLBACK_PORT_TYPE_SOCKS,
						LocalPort: int(port),
						TaskID:    taskData.Task.ID,
					}); err != nil {
						logging.LogError(err, "Failed to start socks")
						response.Error = err.Error()
						response.Success = false
						return response
					} else if !socksResponse.Success {
						response.Error = socksResponse.Error
						response.Success = false
						return response
					} else {
						return response
					}
				} else if action == "stop" {
					if socksResponse, err := mythicrpc.SendMythicRPCProxyStop(mythicrpc.MythicRPCProxyStopMessage{
						PortType: rabbitmq.CALLBACK_PORT_TYPE_SOCKS,
						Port:     int(port),
						TaskID:   taskData.Task.ID,
					}); err != nil {
						logging.LogError(err, "Failed to stop socks")
						response.Error = err.Error()
						response.Success = false
						return response
					} else if !socksResponse.Success {
						response.Error = socksResponse.Error
						response.Success = false
						return response
					} else {
						return response
					}
				} else {
					response.Success = true
					output := "reset all connections and flush data"
					response.DisplayParams = &output
					return response
				}

			}
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return args.LoadArgsFromDictionary(input)
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			return args.LoadArgsFromJSONString(input)
		},
	})
}
