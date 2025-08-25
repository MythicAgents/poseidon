package agentfunctions

import (
	"fmt"
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
				DefaultValue:     -1,
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_NUMBER,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     2,
					},
				},
				Description: "Percentage of jitter on the interval",
			},
			{
				Name:             "backoff_delay",
				ModalDisplayName: "Backoff Delay",
				DefaultValue:     5,
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_NUMBER,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     3,
					},
				},
				Description: "Number of seconds at sleep 0 with no meaningful content before implicitly sleeping to backoff_seconds seconds",
			},
			{
				Name:             "backoff_seconds",
				ModalDisplayName: "Backoff Seconds",
				DefaultValue:     1,
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_NUMBER,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     4,
					},
				},
				Description: "Number of seconds to sleep between checkins if Backoff Delay is triggered while at sleep 0",
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			display := ""
			interval, err := taskData.Args.GetNumberArg("interval")
			if err == nil && interval >= 0 {
				display += fmt.Sprintf("-interval %d ", int(interval))
			}
			jitter, err := taskData.Args.GetNumberArg("jitter")
			if err == nil && jitter >= 0 {
				display += fmt.Sprintf("-jitter %d ", int(jitter))
			}
			backoffDelay, err := taskData.Args.GetNumberArg("backoff_delay")
			if err == nil && backoffDelay >= 0 {
				display += fmt.Sprintf("-backoff_delay %d ", int(backoffDelay))
			}
			backoffSeconds, err := taskData.Args.GetNumberArg("backoff_seconds")
			if err == nil && backoffSeconds >= 0 {
				display += fmt.Sprintf("-backoff_seconds %d ", int(backoffSeconds))
			}
			response.DisplayParams = &display
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
			if len(stringPieces) > 0 {
				if interval, err := strconv.Atoi(stringPieces[0]); err != nil {
					logging.LogError(err, "Failed to process 1st argument as integer")
					return err
				} else if interval < 0 {
					args.SetArgValue("interval", 0)
				} else {
					args.SetArgValue("interval", interval)
				}
			}
			if len(stringPieces) > 1 {
				if interval, err := strconv.Atoi(stringPieces[1]); err != nil {
					logging.LogError(err, "Failed to process 2nd argument as integer")
					return err
				} else if interval < 0 {
					args.SetArgValue("jitter", 0)
				} else {
					args.SetArgValue("jitter", interval)
				}
			}
			if len(stringPieces) > 2 {
				if interval, err := strconv.Atoi(stringPieces[2]); err != nil {
					logging.LogError(err, "Failed to process 3rd argument as integer")
					return err
				} else if interval < 0 {
					args.SetArgValue("backoff_delay", 0)
				} else {
					args.SetArgValue("backoff_delay", interval)
				}
			}
			if len(stringPieces) > 3 {
				if interval, err := strconv.Atoi(stringPieces[3]); err != nil {
					logging.LogError(err, "Failed to process 4th argument as integer")
					return err
				} else if interval < 0 {
					args.SetArgValue("backoff_seconds", 0)
				} else {
					args.SetArgValue("backoff_seconds", interval)
				}
			}
			return nil
		},
	})
}
