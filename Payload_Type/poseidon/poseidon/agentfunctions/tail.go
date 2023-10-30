package agentfunctions

import (
	"fmt"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "tail",
		Description:         "Read the last X lines from a file",
		HelpString:          "tail -path file.txt -lines 5",
		Version:             1,
		Author:              "@its_a_feature_",
		MitreAttackMappings: []string{"T1115"},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "lines",
				ModalDisplayName: "Number of lines to read",
				DefaultValue:     -1,
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_NUMBER,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
					},
				},
				Description: "Number of lines to read from the end of a file",
			},
			{
				Name:             "path",
				ModalDisplayName: "Path to the file to read",
				DefaultValue:     "",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
					},
				},
				Description: "Path to the file to read",
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			lines, err := taskData.Args.GetNumberArg("lines")
			if err != nil {
				response.Error = err.Error()
				response.Success = false
				return response
			}
			path, err := taskData.Args.GetStringArg("path")
			if err != nil {
				response.Error = err.Error()
				response.Success = false
				return response
			}
			displayParams := fmt.Sprintf("%d lines from %s", int(lines), path)
			response.DisplayParams = &displayParams
			return response
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return args.LoadArgsFromDictionary(input)
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			return args.LoadArgsFromJSONString(input)
		},
	})
}
