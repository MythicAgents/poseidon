package agentfunctions

import (
	"encoding/base64"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "jxa",
		Description:         "Execute JavaScript for Automation (JXA) code",
		HelpString:          "jxa {code to execute}",
		Version:             1,
		Author:              "@xorrior",
		MitreAttackMappings: []string{"T1059.002"},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{agentstructs.SUPPORTED_OS_MACOS},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "code",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "JXA Code to execute",
				ModalDisplayName: "JXA Code",
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			return response
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			if err := args.LoadArgsFromDictionary(input); err != nil {
				logging.LogError(err, "Failed to load arguments from dictionary")
				return err
			} else if code, err := args.GetStringArg("code"); err != nil {
				logging.LogError(err, "Failed to get code argument")
				return err
			} else {
				base64Data := base64.StdEncoding.EncodeToString([]byte(code))
				args.SetArgValue("code", base64Data)
				return nil
			}
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			var code string
			if err := args.LoadArgsFromJSONString(input); err != nil {
				code = args.GetCommandLine()
			} else if argCode, err := args.GetStringArg("code"); err != nil {
				logging.LogError(err, "Failed to get code argument from JSON string")
				return err
			} else {
				code = argCode
			}
			base64Data := base64.StdEncoding.EncodeToString([]byte(code))
			args.SetArgValue("code", base64Data)
			return nil
		},
	})
}
