package agentfunctions

import (
	"errors"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
	"strconv"
	"strings"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "sleep",
		Description:         "Update the sleep interval of the agent.",
		HelpString:          "sleep {interval} [jitter%]",
		Version:             1,
		Author:              "@xorrior",
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "interval",
				ModalDisplayName: "Interval Seconds",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_NUMBER,
				DefaultValue:     0,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
					},
				},
				Description: "Sleep time in seconds",
			},
			{
				Name:             "jitter",
				ModalDisplayName: "Jitter Percentage",
				DefaultValue:     0,
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_NUMBER,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     2,
					},
				},
				Description: "Percentage of jitter on the interval",
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			return response
		},
		TaskFunctionProcessResponse: func(processResponse agentstructs.PtTaskProcessResponseMessage) agentstructs.PTTaskProcessResponseMessageResponse {
			response := agentstructs.PTTaskProcessResponseMessageResponse{
				TaskID:  processResponse.TaskData.Task.ID,
				Success: true,
			}
			sleepString := processResponse.Response.(string)
			if updateResp, err := mythicrpc.SendMythicRPCCallbackUpdate(mythicrpc.MythicRPCCallbackUpdateMessage{
				AgentCallbackUUID: &processResponse.TaskData.Callback.AgentCallbackID,
				SleepInfo:         &sleepString,
			}); err != nil {
				response.Success = false
				response.Error = err.Error()
			} else if !updateResp.Success {
				response.Success = false
				response.Error = updateResp.Error
			}
			return response
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return args.LoadArgsFromDictionary(input)
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			stringPieces := strings.Split(input, "")
			if len(stringPieces) == 1 {
				if interval, err := strconv.Atoi(stringPieces[0]); err != nil {
					logging.LogError(err, "Failed to process first argument as integer")
					return err
				} else if interval < 0 {
					args.SetArgValue("interval", 0)
				} else {
					args.SetArgValue("interval", interval)
				}
				return nil
			} else if len(stringPieces) == 2 {
				if interval, err := strconv.Atoi(stringPieces[0]); err != nil {
					return err
				} else if jitter, err := strconv.Atoi(stringPieces[1]); err != nil {
					return err
				} else {
					if interval < 0 {
						args.SetArgValue("interval", 0)
					} else {
						args.SetArgValue("interval", interval)
					}
					if jitter < 0 {
						args.SetArgValue("jitter", 0)
					} else {
						args.SetArgValue("jitter", jitter)
					}
					return nil
				}
			} else {
				return errors.New("Too many arguments, expecting two")
			}
		},
	})
}
