package agentfunctions

import (
	"errors"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                  "libinject",
		Description:           "Inject a library from on-host into a process.",
		HelpString:            "libinject",
		Version:               1,
		Author:                "@xorrior",
		MitreAttackMappings:   []string{"T1055"},
		SupportedUIFeatures:   []string{},
		NeedsAdminPermissions: true,
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{agentstructs.SUPPORTED_OS_MACOS},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "pid",
				ModalDisplayName: "PID to inject into",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_NUMBER,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
					},
				},
				Description: "PID to inject a library into",
			},
			{
				Name:             "library",
				ModalDisplayName: "Library Path on Disk",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
					},
				},
				Description: "Absolute path to dylib on target to load",
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			if taskData.Callback.IntegrityLevel <= 2 {
				response.Success = false
				response.Error = "Must be elevated to run this command"
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
