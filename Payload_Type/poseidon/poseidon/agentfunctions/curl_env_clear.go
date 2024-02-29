package agentfunctions

import (
	"errors"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "curl_env_clear",
		Description:         "Clear environment variables to use in subsequent curl requests",
		HelpString:          "curl_env_clear -clearEnv TOKEN -clearEnv URL",
		Version:             1,
		Author:              "@its_a_feature_",
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "clearEnv",
				ModalDisplayName: "Environment Names to clear",
				DefaultValue:     []string{},
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_ARRAY,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     1,
					},
				},
				Description: "Array of environment names to clear",
			},
			{
				Name:             "clearAllEnv",
				ModalDisplayName: "Clear All Environment variables",
				DefaultValue:     false,
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     3,
					},
				},
				Description: "Clear all environment variables",
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			clearEnvList, err := taskData.Args.GetArrayArg("clearEnv")
			if err != nil {
				logging.LogError(err, "Failed to get url string")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			if len(clearEnvList) == 0 {
				taskData.Args.SetArgValue("clearAllEnv", true)
			}
			commandName := "curl"
			response.CommandName = &commandName
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
