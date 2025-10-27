package agentfunctions

import (
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
	"github.com/MythicMeta/MythicContainer/utils/helpers"
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
						GroupName:           "Default",
						UIModalPosition:     1,
					},
				},
			},
			{
				Name:                 "existingFile",
				ModalDisplayName:     "Existing File",
				ParameterType:        agentstructs.COMMAND_PARAMETER_TYPE_CHOOSE_ONE,
				Description:          "Name of an existing file to upload",
				Choices:              []string{""},
				DefaultValue:         "",
				DynamicQueryFunction: getUploadFiles,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						GroupName:           "existingFile",
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
						GroupName:           "Default",
						UIModalPosition:     2,
					},
					{
						ParameterIsRequired: false,
						GroupName:           "existingFile",
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
						GroupName:           "Default",
						UIModalPosition:     3,
					},
					{
						ParameterIsRequired: false,
						GroupName:           "existingFile",
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
			var search *mythicrpc.MythicRPCFileSearchMessageResponse
			groupName, err := taskData.Args.GetParameterGroupName()
			if err != nil {
				logging.LogError(err, "failed to get group name")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			if groupName == "Default" {
				fileID, err := taskData.Args.GetFileArg("file_id")
				if err != nil {
					logging.LogError(err, "Failed to get file_id")
					response.Success = false
					response.Error = err.Error()
					return response
				}
				search, err = mythicrpc.SendMythicRPCFileSearch(mythicrpc.MythicRPCFileSearchMessage{
					AgentFileID: fileID,
				})
				taskData.Args.RemoveArg("file_id")
			} else {
				filename, err := taskData.Args.GetStringArg("existingFile")
				if err != nil {
					logging.LogError(err, "Failed to get file_id")
					response.Success = false
					response.Error = err.Error()
					return response
				}
				search, err = mythicrpc.SendMythicRPCFileSearch(mythicrpc.MythicRPCFileSearchMessage{
					Filename:   filename,
					TaskID:     taskData.Task.ID,
					MaxResults: 1,
				})
				taskData.Args.RemoveArg("existingFile")
			}
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			if !search.Success {
				response.Success = false
				response.Error = search.Error
				return response
			}
			if len(search.Files) == 0 {
				response.Success = false
				response.Error = "Failed to find the specified file, was it deleted?"
				return response
			}
			taskData.Args.AddArg(agentstructs.CommandParameter{
				Name:         "file_id",
				DefaultValue: search.Files[0].AgentFileId,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						GroupName: groupName,
					},
				},
			})
			remotePath, err := taskData.Args.GetStringArg("remote_path")
			if err != nil {
				logging.LogError(err, "Failed to get remote path parameter")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			if len(remotePath) == 0 {
				// set the remote path to just the filename to upload it to the same directory our agent is in
				taskData.Args.SetArgValue("remote_path", search.Files[0].Filename)
				displayString := fmt.Sprintf("%s",
					search.Files[0].Filename)
				response.DisplayParams = &displayString
				return response
			}
			displayString := fmt.Sprintf("%s",
				search.Files[0].Filename)
			response.DisplayParams = &displayString
			return response

		},
	})
}
func getUploadFiles(input agentstructs.PTRPCDynamicQueryFunctionMessage) []string {
	fileResp, err := mythicrpc.SendMythicRPCFileSearch(mythicrpc.MythicRPCFileSearchMessage{
		LimitByCallback:     false,
		CallbackID:          input.Callback,
		IsPayload:           false,
		IsDownloadFromAgent: false,
		Filename:            "",
	})
	if err != nil {
		logging.LogError(err, "Failed to search for files in callback")
		return []string{}
	}
	if !fileResp.Success {
		logging.LogError(err, "Failed to search for files in callback", "mythic error", fileResp.Error)
		return []string{}
	}
	potentialFiles := []string{}
	for _, file := range fileResp.Files {
		if !helpers.StringSliceContains(potentialFiles, file.Filename) {
			potentialFiles = append(potentialFiles, file.Filename)
		}
	}
	return potentialFiles

}
