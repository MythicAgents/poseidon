package agentfunctions

import (
	"errors"
	"fmt"
	"github.com/MythicMeta/MythicContainer/logging"
	"strconv"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "link_tcp",
		Description:         "Link one poseidon agent to another over poseidon_tcp.",
		HelpString:          "link_tcp {IP | Host} {port}",
		Version:             1,
		Author:              "@its_a_feature_",
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:          "address",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
						GroupName:           "Default",
					},
				},
				Description: "Address of the computer to connect to (IP or Hostname)",
			},
			{
				Name:          "port",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_NUMBER,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
						GroupName:           "Default",
					},
				},
				Description: "Port to connect to that the remote agent is listening on",
			},
			{
				Name:          "connection",
				CLIName:       "connectionDictionary",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_CONNECTION_INFO,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
						GroupName:           "Mythic Modal",
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
			groupName, err := taskData.Args.GetParameterGroupName()
			if err != nil {
				logging.LogError(err, "Failed to get parameter group name")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			if groupName == "Default" {
				address, err := taskData.Args.GetStringArg("address")
				if err != nil {
					response.Error = err.Error()
					response.Success = false
					return response
				}
				port, err := taskData.Args.GetNumberArg("port")
				if err != nil {
					response.Error = err.Error()
					response.Success = false
					return response
				}
				displayString := fmt.Sprintf("%s on port %.0f", address, port)
				response.DisplayParams = &displayString

			} else {
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
				err = taskData.Args.SetArgValue("address", connectionInfo.Host)
				if err != nil {
					logging.LogError(err, "Failed to get address information")
					response.Success = false
					response.Error = err.Error()
					return response
				}
				port, err := strconv.Atoi(connectionInfo.C2ProfileInfo.Parameters["port"].(string))
				if err != nil {
					logging.LogError(err, "Failed to convert port to integer")
					response.Success = false
					response.Error = err.Error()
					return response
				}
				err = taskData.Args.SetArgValue("port", port)
				if err != nil {
					logging.LogError(err, "Failed to get port information")
					response.Success = false
					response.Error = err.Error()
					return response
				}
				displayString := fmt.Sprintf("%s on port %d", connectionInfo.Host, port)
				response.DisplayParams = &displayString
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
