package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "curl_env_get",
		Description:         "Get environment variables to use in subsequent curl requests",
		HelpString:          "curl_env_get",
		Version:             1,
		Author:              "@its_a_feature_",
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		CommandParameters: []agentstructs.CommandParameter{},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			taskData.Args.AddArg(agentstructs.CommandParameter{Name: "getEnv", DefaultValue: true, ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN})
			commandName := "curl"
			response.CommandName = &commandName
			return response
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return nil
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			return nil
		},
	})
}
