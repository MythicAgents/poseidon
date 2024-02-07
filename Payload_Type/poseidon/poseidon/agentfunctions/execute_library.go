package agentfunctions

import (
	"fmt"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
	"strings"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "execute_library",
		HelpString:          "execute_library",
		Description:         "Load a dylib from disk and run a function within it.",
		Version:             1,
		MitreAttackMappings: []string{"T1106", "T1620", "T1105"},
		Author:              "@its_a_feature_",
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:          "function_name",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:   "Which function should be executed?",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
						GroupName:           "New File",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
						GroupName:           "Existing File",
					},
				},
			},
			{
				Name:          "file_path",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:   "Where is the dylib on disk to load up or where should the uploaded one be written to?",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
						GroupName:           "New File",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
						GroupName:           "Existing File",
					},
				},
			},
			{
				Name:             "file_id",
				ModalDisplayName: "Binary/Bundle to execute",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_FILE,
				Description:      "Select the Bundle/Dylib/Binary to execute in memory",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
						GroupName:           "New File",
					},
				},
			},
			{
				Name:             "args",
				ModalDisplayName: "Arguments",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_ARRAY,
				Description:      "Arguments to pass to function",
				DefaultValue:     []string{},
				//Choices:          []string{"int*", "char*"},
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     3,
						GroupName:           "New File",
					},
					{
						ParameterIsRequired: true,
						UIModalPosition:     3,
						GroupName:           "Existing File",
					},
				},
				TypedArrayParseFunction: func(message agentstructs.PTRPCTypedArrayParseFunctionMessage) [][]string {
					responseArray := [][]string{}
					for _, msg := range message.InputArray {
						inputPieces := strings.Split(msg, ":")
						if len(inputPieces) < 2 {
							logging.LogError(nil, "Failed to parse typed array", "element", msg)
							continue
						}
						value := strings.Join(inputPieces[1:], ":")
						if len(value) > 0 {
							if value[0] == '"' && value[len(value)-1] == '"' {
								value = value[1 : len(value)-1]
							}
						}
						responseArray = append(responseArray, []string{inputPieces[0], value})
					}
					return responseArray
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
			groupName, err := taskData.Args.GetParameterGroupName()
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			if groupName == "New File" {
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
					Comment:     "Uploaded to disk for execute_library",
				}); err != nil {
					response.Success = false
					response.Error = err.Error()
					return response
				} else if funcName, err := taskData.Args.GetStringArg("function_name"); err != nil {
					response.Success = false
					response.Error = err.Error()
					return response
				} else {
					displayString := fmt.Sprintf("function %s of %s",
						funcName, search.Files[0].Filename)
					response.DisplayParams = &displayString
					return response
				}
			} else {
				return response
			}
		},
	})
}
