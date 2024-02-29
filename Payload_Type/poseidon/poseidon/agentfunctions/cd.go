package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"strings"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "cd",
		Description:         "Change working directory (can be relative, but no ~).",
		HelpString:          "cd -path [new directory]",
		Version:             1,
		Author:              "@xorrior, @its_a_feature_",
		MitreAttackMappings: []string{"T1005"},
		SupportedUIFeatures: []string{},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:          "path",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:  "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
					},
				},
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			displayParams, err := taskData.Args.GetStringArg("path")
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			response.DisplayParams = &displayParams
			taskData.Args.SetManualArgs(displayParams)
			return response
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			err := args.LoadArgsFromDictionary(input)
			if err != nil {
				return err
			}
			path, err := args.GetStringArg("path")
			if err != nil {
				return err
			}
			args.SetArgValue("path", strings.Trim(path, "\""))
			return nil
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			err := args.LoadArgsFromJSONString(input)
			if err != nil {
				args.SetArgValue("path", strings.Trim(input, "\""))
			}
			return nil
		},
	})
}
