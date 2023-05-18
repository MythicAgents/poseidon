package agentfunctions

import (
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "upload",
		HelpString:          "upload",
		Description:         "Upload a file to the target",
		Version:             1,
		MitreAttackMappings: []string{"T1020", "T1030", "T1041", "T1105"},
		Author:              "@xorrior",
		SupportedUIFeatures: []string{"file_browser:upload"},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "file_id",
				ModalDisplayName: "File to Upload",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_FILE,
				Description:      "Select a file to write to the remote path",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
					},
				},
			},
			{
				Name:             "remote_path",
				ModalDisplayName: "Remote Path",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Path where the uploaded file will be written",
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     2,
					},
				},
			},
			{
				Name:             "overwrite",
				ModalDisplayName: "Overwrite existing file",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				Description:      "Overwrite file if it exists",
				DefaultValue:     false,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     3,
					},
				},
			},
		},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			return args.LoadArgsFromJSONString(input)
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return args.LoadArgsFromDictionary(input)
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			if fileID, err := taskData.Args.GetFileArg("file_id"); err != nil {
				logging.LogError(err, "Failed to get file_id")
				response.Success = false
				response.Error = err.Error()
				return response
			} else if search, err := mythicrpc.SendMythicRPCFileSearch(mythicrpc.MythicRPCFileSearchMessage{
				AgentFileID: fileID,
			}); err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			} else if !search.Success {
				response.Success = false
				response.Error = search.Error
				return response
			} else if len(search.Files) == 0 {
				response.Success = false
				response.Error = "Failed to find the specified file, was it deleted?"
				return response
			} else if remotePath, err := taskData.Args.GetStringArg("remote_path"); err != nil {
				logging.LogError(err, "Failed to get remote path parameter")
				response.Success = false
				response.Error = err.Error()
				return response
			} else if len(remotePath) == 0 {
				// set the remote path to just the filename to upload it to the same directory our agent is in
				taskData.Args.SetArgValue("remote_path", search.Files[0].Filename)
				displayString := fmt.Sprintf("%s",
					search.Files[0].Filename)
				response.DisplayParams = &displayString
				return response
			} else {
				displayString := fmt.Sprintf("%s",
					search.Files[0].Filename)
				response.DisplayParams = &displayString
				return response
			}
		},
	})
}
