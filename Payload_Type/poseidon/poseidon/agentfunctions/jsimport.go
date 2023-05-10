package agentfunctions

import (
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "jsimport",
		HelpString:          "jsimport",
		Description:         "Upload a script into memory for use with jsimport_call",
		Version:             1,
		MitreAttackMappings: []string{"T1020", "T1030", "T1041", "T1620", "T1105"},
		Author:              "@its_a_feature_",
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "file_id",
				ModalDisplayName: "JXA Script to Load",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_FILE,
				Description:      "Select the JXA Script to load into memory",
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
			if fileID, err := taskData.Args.GetStringArg("file_id"); err != nil {
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
			} else if _, err := mythicrpc.SendMythicRPCFileUpdate(mythicrpc.MythicRPCFileUpdateMessage{
				AgentFileID: fileID,
				Comment:     "Uploaded into memory for jsimport",
			}); err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			} else {
				displayString := fmt.Sprintf("script %s",
					search.Files[0].Filename)
				response.DisplayParams = &displayString
				return response
			}
		},
	})
}
