package agentfunctions

import (
	"errors"
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"path/filepath"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "download_bulk",
		HelpString:          "download_bulk -paths /Users/bob/Desktop -paths /Users/bob/Downloads -compress",
		Description:         "Download file(s), optionally compressing into a Zip before download. Stored in memory prior to upload - may be resource intensive.",
		Version:             1,
		MitreAttackMappings: []string{"T1020", "T1030", "T1041"},
		Author:              "@maclarel",
		AssociatedBrowserScript: &agentstructs.BrowserScript{
			ScriptPath: filepath.Join(".", "poseidon", "browserscripts", "download_bulk.js"),
			Author:     "@maclarel",
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "paths",
				ModalDisplayName: "Remote Path(s)",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_ARRAY,
				Description:      "Paths of file(s) to retrieve",
				DefaultValue:     []string{},
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
					},
				},
			},
			{
				Name:             "compress",
				ModalDisplayName: "Compress file(s)",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				Description:      "Compress files prior to transfer",
				DefaultValue:     true,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     2,
					},
				},
			},
		},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			displayParams := ""
			paths, err := taskData.Args.GetArrayArg("paths")
			if err != nil {
				logging.LogError(err, "failed to get paths argument")
				response.Success = false
				return response
			}
			for _, path := range paths {
				displayParams += fmt.Sprintf(" -path \"%s\"", path)
			}
			compress, err := taskData.Args.GetBooleanArg("compress")
			if err != nil {
				logging.LogError(err, "failed to get compress")
				response.Success = false
				return response
			}
			if compress {
				displayParams += " -compress"
			}
			response.DisplayParams = &displayParams
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
