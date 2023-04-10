package agentfunctions

import (
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "execute_macho",
		HelpString:          "execute_macho",
		Description:         "Upload a thin x64_Mach-o binary into memory and execute a function in-proc",
		Version:             1,
		MitreAttackMappings: []string{"T1106", "T1620", "T1105"},
		Author:              "",
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:          "args",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_ARRAY,
				Description:   "Arguments to pass to binary",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
					},
				},
			},
			{
				Name:             "file_id",
				ModalDisplayName: "Binary to execute",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_FILE,
				Description:      "Select the Binary to execute in memory",
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
				Comment:     "Uploaded into memory for execute_macho",
			}); err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			} else if binaryArgs, err := taskData.Args.GetStringArg("args"); err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			} else {
				displayString := fmt.Sprintf("%s with args %s",
					search.Files[0].Filename, binaryArgs)
				response.DisplayParams = &displayString
				return response
			}
		},
	})
}
