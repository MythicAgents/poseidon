package agentfunctions

import (
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "unlink_tcp",
		Description:         "Unlink a poseidon_tcp connection.",
		HelpString:          "unlink_tcp",
		Version:             1,
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{},
		Author:              "@its_a_feature_",
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:          "connection",
				Description:   "Connection info for unlinking",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_LINK_INFO,
			},
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			return nil
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return args.LoadArgsFromDictionary(input)
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			if connectionInfo, err := taskData.Args.GetLinkInfoArg("connection"); err != nil {
				response.Success = false
				response.Error = err.Error()
			} else if connectionInfo.CallbackUUID == "" {
				response.Success = false
				response.Error = "Failed to find callback UUID in connection information"
			} else {
				taskData.Args.RemoveArg("connection")
				taskData.Args.AddArg(agentstructs.CommandParameter{
					Name:          "connection",
					DefaultValue:  connectionInfo.CallbackUUID,
					ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				})
				displayString := fmt.Sprintf("from %s", connectionInfo.CallbackUUID)
				response.DisplayParams = &displayString
			}
			return response
		},
	})
}
