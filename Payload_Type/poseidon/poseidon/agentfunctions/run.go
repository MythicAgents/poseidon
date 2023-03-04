package agentfunctions

import (
	"errors"
	"github.com/MythicMeta/MythicContainer/logging"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "run",
		Description:         "Execute a command from disk with arguments.",
		HelpString:          "run -path /path/to/binary -args arg1 -args arg2 -args arg3",
		Version:             1,
		Author:              "@its_a_feature_",
		MitreAttackMappings: []string{"T1059.004"},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "path",
				ModalDisplayName: "Binary Path",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
					},
				},
				Description: "Absolute path to the program to run",
			},
			{
				Name:             "args",
				ModalDisplayName: "Arguments",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_ARRAY,
				DefaultValue:     []string{},
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     2,
					},
				},
				Description: "Array of arguments to pass to the program.",
			},
		},
		TaskFunctionCreateTasking: func(taskData agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			if path, err := taskData.Args.GetStringArg("path"); err != nil {
				logging.LogError(err, "Failed to get path argument")
				response.Success = false
				response.Error = err.Error()
				return response
			} else {
				response.DisplayParams = &path
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
