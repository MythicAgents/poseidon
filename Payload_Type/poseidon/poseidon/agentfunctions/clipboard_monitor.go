package agentfunctions

import (
	"fmt"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "clipboard_monitor",
		Description:         "Monitor the macOS clipboard for changes every X seconds",
		HelpString:          "clipboard_monitor -duration -1",
		Version:             1,
		Author:              "@its_a_feature_",
		MitreAttackMappings: []string{"T1115"},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{agentstructs.SUPPORTED_OS_MACOS},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "duration",
				ModalDisplayName: "Monitor Duration",
				DefaultValue:     -1,
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_NUMBER,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     1,
					},
				},
				Description: "Number of seconds to monitor the clipboard, or a negative value to do it indefinitely",
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			if duration, err := taskData.Args.GetNumberArg("duration"); err != nil {
				logging.LogError(err, "Failed to get duration during create tasking")
				response.Success = false
				response.Error = err.Error()
				return response
			} else if duration < 0 {
				displayParams := "indefinitely"
				response.DisplayParams = &displayParams
			} else {
				displayParams := fmt.Sprintf("for %02f seconds", duration)
				response.DisplayParams = &displayParams
			}
			return response
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return args.LoadArgsFromDictionary(input)
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			if len(input) > 0 {
				if err := args.LoadArgsFromJSONString(input); err != nil {
					args.SetArgValue("duration", input)
					return nil
				} else {
					return nil
				}
			} else {
				return args.SetArgValue("duration", -1)
			}
		},
	})
}
