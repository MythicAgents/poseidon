package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/mitchellh/mapstructure"
	"path/filepath"
)

var download = agentstructs.Command{
	Name:                "download",
	HelpString:          "download [path]",
	Description:         "Download a file from the target",
	Version:             1,
	MitreAttackMappings: []string{"T1020", "T1030", "T1041"},
	SupportedUIFeatures: []string{"file_browser:download"},
	AssociatedBrowserScript: &agentstructs.BrowserScript{
		ScriptPath: filepath.Join(".", "poseidon", "browserscripts", "download_new.js"), // the name of the script in agent_browser_scripts
		Author:     "@its_a_feature_",
	},
	TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
		response := agentstructs.PTTaskCreateTaskingMessageResponse{
			Success: true,
			TaskID:  taskData.Task.ID,
		}
		if displayParams, err := taskData.Args.GetFinalArgs(); err != nil {
			logging.LogError(err, "Failed to get final arguments for task")
			response.Success = false
			response.Error = err.Error()
			return response
		} else {
			response.DisplayParams = &displayParams
		}
		return response
	},
	TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
		//return args.LoadArgsFromDictionary(input)
		fileBrowserData := agentstructs.FileBrowserTask{}
		if err := mapstructure.Decode(input, &fileBrowserData); err != nil {
			logging.LogError(err, "Failed to marshal file browser data")
			return err
		} else {
			// manually set the arguments to be the full path to the thing we want to download
			args.SetManualArgs(fileBrowserData.FullPath)
			return nil
		}
	},
	TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
		//return args.LoadArgsFromJSONString(input)
		return nil
	},
}

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(download)
}
