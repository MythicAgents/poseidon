package agentfunctions

import (
	"errors"
	"fmt"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "persist_loginitem",
		Description:         "Add a login item for the current user via the LSSharedFileListInsertItemURL function",
		HelpString:          "persist_loginitem",
		Version:             1,
		Author:              "@xorrior, @its_a_feature_",
		MitreAttackMappings: []string{"T1547.015", "T1647"},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{agentstructs.SUPPORTED_OS_MACOS},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "path",
				ModalDisplayName: "Program Location",
				DefaultValue:     "",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     1,
					},
				},
				Description: "Path to the binary to execute at login",
			},
			{
				Name:          "name",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:  "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     2,
					},
				},
				Description: "The name that is displayed in the Login Items section of the Users & Groups preferences pane",
			},
			{
				Name:          "global",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				DefaultValue:  false,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     3,
					},
				},
				Description: "Set this to true if the login item should be installed for all users. This requires administrative privileges",
			},
			{
				Name:          "list",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				DefaultValue:  false,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     4,
					},
				},
				Description: "List current global and session items",
			},
			{
				Name:          "remove",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				DefaultValue:  false,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     5,
					},
				},
				Description: "Remove the specified login item by path and name",
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			path, err := taskData.Args.GetStringArg("path")
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			name, err := taskData.Args.GetStringArg("name")
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			list, err := taskData.Args.GetBooleanArg("list")
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			remove, err := taskData.Args.GetBooleanArg("remove")
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			if list {
				displayString := fmt.Sprintf("listing session and global instances")
				response.DisplayParams = &displayString
			} else if remove {
				displayString := fmt.Sprintf("to remove %s as %s", path, name)
				response.DisplayParams = &displayString
			} else {
				displayString := fmt.Sprintf("to add %s as %s", path, name)
				response.DisplayParams = &displayString
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
