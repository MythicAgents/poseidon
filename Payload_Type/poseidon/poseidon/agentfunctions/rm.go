package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/mitchellh/mapstructure"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "rm",
		Description:         "rm [path]",
		Version:             1,
		MitreAttackMappings: []string{"T1070.004"},
		SupportedUIFeatures: []string{"file_browser:remove"},
		Author:              "@xorrior",
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			if path, err := taskData.Args.GetFinalArgs(); err != nil {
				logging.LogError(err, "Failed to get final args")
				response.Error = err.Error()
				response.Success = false
				return response
			} else {
				response.DisplayParams = &path
			}
			return response
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			// if we get a dictionary, it'll be from the file browser which will supply agentstructs.FileBrowserTask data
			fileBrowserData := agentstructs.FileBrowserTask{}
			//logging.LogDebug("Called TaskFunctionParseArgDictionary in ls")
			if err := mapstructure.Decode(input, &fileBrowserData); err != nil {
				logging.LogError(err, "Failed to get file browser data struct information from dictionary input")
				return err
			} else {
				args.AddArg(agentstructs.CommandParameter{
					Name:          "file",
					DefaultValue:  fileBrowserData.FullPath,
					ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				})
				return nil
			}
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			//return args.LoadArgsFromJSONString(input)
			args.AddArg(agentstructs.CommandParameter{
				Name:          "file",
				DefaultValue:  args.GetCommandLine(),
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
			})
			return nil
		},
	})
}
