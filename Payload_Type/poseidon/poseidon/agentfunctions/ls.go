package agentfunctions

import (
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/mitchellh/mapstructure"
	"path/filepath"
	"strings"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "ls",
		Description:         "List out the contents of a directory with an optional depth flag for recursion",
		HelpString:          "ls -path path -depth 1",
		Version:             1,
		MitreAttackMappings: []string{"T1083"},
		SupportedUIFeatures: []string{"file_browser:list"},
		Author:              "@xorrior",
		AssociatedBrowserScript: &agentstructs.BrowserScript{
			ScriptPath: filepath.Join(".", "poseidon", "browserscripts", "ls_new.js"),
			Author:     "@its_a_feature_",
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:          "path",
				Description:   "Path to a directory",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:  "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     0,
					},
				},
			},
			{
				Name:          "depth",
				Description:   "Depth for recursive directory listings",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_NUMBER,
				DefaultValue:  1,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     1,
					},
				},
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			path, err := taskData.Args.GetStringArg("path")
			if err != nil {
				logging.LogError(err, "Failed to get string arg for path")
				response.Error = err.Error()
				response.Success = false
				return response
			}
			depth, err := taskData.Args.GetNumberArg("depth")
			if err != nil {
				logging.LogError(err, "Failed to get string arg for path")
				response.Error = err.Error()
				response.Success = false
				return response
			}
			displayParams := fmt.Sprintf("-path \"%s\" -depth %.0f", path, depth)
			response.DisplayParams = &displayParams
			return response
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			err := args.LoadArgsFromDictionary(input)
			if err != nil {
				logging.LogError(err, "failed to load dictionary args for ls")
			}
			fileBrowserData := agentstructs.FileBrowserTask{}
			err = mapstructure.Decode(input, &fileBrowserData)
			if err != nil {
				logging.LogError(err, "Failed to get file browser data struct information from dictionary input")
				return err
			}
			if fileBrowserData.Host != "" {
				args.AddArg(agentstructs.CommandParameter{
					Name:          "file_browser",
					ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
					DefaultValue:  true,
				})
			}
			if fileBrowserData.FullPath != "" {
				args.SetArgValue("path", strings.Trim(fileBrowserData.FullPath, "\""))
			}
			path, err := args.GetStringArg("path")
			if err != nil || path == "" {
				args.SetArgValue("path", ".")
			}
			return nil

		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			if input == "" {
				args.AddArg(agentstructs.CommandParameter{
					Name:          "path",
					ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
					DefaultValue:  ".",
				})
			} else {
				args.LoadArgsFromJSONString(input)
				path, err := args.GetStringArg("path")
				if err != nil || path == "" {
					args.SetArgValue("path", ".")
				} else {
					args.AddArg(agentstructs.CommandParameter{
						Name:          "path",
						ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
						DefaultValue:  strings.Trim(input, "\""),
					})
				}
			}
			args.AddArg(agentstructs.CommandParameter{
				Name:          "file_browser",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				DefaultValue:  false,
			})
			return nil
		},
	})
}
