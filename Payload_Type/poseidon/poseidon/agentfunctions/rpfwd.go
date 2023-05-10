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
		Name:                "rpfwd",
		Description:         "Start or Stop a Reverse Port Forward.",
		HelpString:          "rpfwd",
		Version:             1,
		Author:              "@its_a_feature_",
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "action",
				ModalDisplayName: "Action",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_CHOOSE_ONE,
				Choices:          []string{"start", "stop"},
				DefaultValue:     "start",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
						GroupName:           "start",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
						GroupName:           "stop",
					},
				},
				Description: "Start or Stop rpfwd through this callback",
			},
			{
				Name:             "port",
				ModalDisplayName: "Local Port",
				DefaultValue:     7000,
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_NUMBER,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
						GroupName:           "start",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
						GroupName:           "stop",
					},
				},
				Description: "Local port to open on host where agent is running",
			},
			{
				Name:             "remote_port",
				ModalDisplayName: "Remote Port",
				DefaultValue:     7000,
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_NUMBER,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     4,
						GroupName:           "start",
					},
				},
				Description: "Remote port to connect to when a new connection comes in",
			},
			{
				Name:             "remote_ip",
				ModalDisplayName: "Remote IP",
				DefaultValue:     "",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     3,
						GroupName:           "start",
					},
				},
				Description: "Remote IP to connect to when a new connection comes in",
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
			} else if remotePort, err := taskData.Args.GetNumberArg("remote_port"); err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			} else if remoteIP, err := taskData.Args.GetStringArg("remote_ip"); err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			} else {
				if action == "start" {
					displayString := fmt.Sprintf("%s on port %.0f with reverse connection to %s:%.0f", action, port,
						remoteIP, remotePort)
					response.DisplayParams = &displayString
					if socksResponse, err := mythicrpc.SendMythicRPCProxyStart(mythicrpc.MythicRPCProxyStartMessage{
						PortType:   rabbitmq.CALLBACK_PORT_TYPE_RPORTFWD,
						LocalPort:  int(port),
						RemotePort: int(remotePort),
						RemoteIP:   remoteIP,
						TaskID:     taskData.Task.ID,
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

						taskData.Args.RemoveArg("remote_port")
						taskData.Args.RemoveArg("remote_ip")
						taskData.Args.SetManualParameterGroup("start")
						return response
					}
				} else {
					displayString := fmt.Sprintf("%s on port %.0f", action, port)
					response.DisplayParams = &displayString
					if socksResponse, err := mythicrpc.SendMythicRPCProxyStop(mythicrpc.MythicRPCProxyStopMessage{
						PortType: rabbitmq.CALLBACK_PORT_TYPE_RPORTFWD,
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
