package agentfunctions

import (
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
	"github.com/MythicMeta/MythicContainer/rabbitmq"
)

var pty = agentstructs.Command{
	Name:                      "pty",
	Description:               "open up an interactive pty",
	MitreAttackMappings:       []string{"T1059"},
	TaskFunctionCreateTasking: ptyCreateTasking,
	SupportedUIFeatures:       []string{"task_response:interactive"},
	CommandParameters: []agentstructs.CommandParameter{
		{
			Name:             "program_path",
			CLIName:          "program_path",
			ModalDisplayName: "Program Path",
			ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
			Description:      "What program to spawn with a PTY",
			DefaultValue:     "/bin/bash",
		},
		{
			Name:             "open_port",
			CLIName:          "openPort",
			ModalDisplayName: "Open Port",
			ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
			Description:      "Whether to open a local port for additional PTY access",
			DefaultValue:     false,
			ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
				{
					ParameterIsRequired: false,
				},
			},
		},
	},
	TaskCompletionFunctions: map[string]agentstructs.PTTaskCompletionFunction{
		"close_ports": func(taskData *agentstructs.PTTaskMessageAllData, subtaskData *agentstructs.PTTaskMessageAllData, subtaskName *agentstructs.SubtaskGroupName) agentstructs.PTTaskCompletionFunctionMessageResponse {
			response := agentstructs.PTTaskCompletionFunctionMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			if socksResponse, err := mythicrpc.SendMythicRPCProxyStop(mythicrpc.MythicRPCProxyStopMessage{
				PortType: rabbitmq.CALLBACK_PORT_TYPE_INTERACTIVE,
				Port:     0,
				TaskID:   taskData.Task.ID,
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
				stdout := fmt.Sprintf("Closed port: %d\n", socksResponse.LocalPort)
				response.Stdout = &stdout
				return response
			}
		},
	},
	TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
		return args.LoadArgsFromDictionary(input)
	},
	Version: 1,
}

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(pty)
}

func ptyCreateTasking(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
	response := agentstructs.PTTaskCreateTaskingMessageResponse{
		Success: true,
		TaskID:  taskData.Task.ID,
	}
	programPath, err := taskData.Args.GetStringArg("program_path")
	if err != nil {
		response.Error = err.Error()
		response.Success = false
		return response
	}
	_, err = mythicrpc.SendMythicRPCArtifactCreate(mythicrpc.MythicRPCArtifactCreateMessage{
		BaseArtifactType: "ProcessCreate",
		ArtifactMessage:  programPath,
		TaskID:           taskData.Task.ID,
	})
	if err != nil {
		logging.LogError(err, "Failed to send mythicrpc artifact create")
	}
	openPort, err := taskData.Args.GetBooleanArg("open_port")
	if err != nil {
		response.Error = err.Error()
		response.Success = false
		return response
	}
	if openPort {
		socksResponse, err := mythicrpc.SendMythicRPCProxyStart(mythicrpc.MythicRPCProxyStartMessage{
			PortType:  rabbitmq.CALLBACK_PORT_TYPE_INTERACTIVE,
			LocalPort: 0,
			TaskID:    taskData.Task.ID,
		})
		if err != nil {
			logging.LogError(err, "Failed to start socks")
			response.Error = err.Error()
			response.Success = false
			return response
		}
		if !socksResponse.Success {
			response.Error = socksResponse.Error
			response.Success = false
			return response
		}
		stdout := fmt.Sprintf("Opened port: %d\n", socksResponse.LocalPort)
		response.Stdout = &stdout
		completionName := "close_ports"
		response.CompletionFunctionName = &completionName
	}
	response.DisplayParams = &programPath
	return response
}
