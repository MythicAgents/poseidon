package agentfunctions

import (
	"encoding/base64"
	"errors"
	"fmt"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "curl",
		Description:         "Execute a single web request",
		HelpString:          "curl -url https://www.google.com -method GET -headers \"Host: abc.com\" -headers \"Authorization: Bearer $TOKEN\"",
		Version:             1,
		Author:              "@xorrior",
		MitreAttackMappings: []string{"T1213"},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:          "url",
				DefaultValue:  "https://www.google.com",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
					},
				},
				Description: "URL to request",
			},
			{
				Name:             "method",
				ModalDisplayName: "HTTP Method",
				DefaultValue:     "GET",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_CHOOSE_ONE,
				Choices:          []string{"GET", "POST", "PUT", "DELETE"},
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     2,
					},
				},
				Description: "Type of request",
			},
			{
				Name:             "headers",
				ModalDisplayName: "HTTP Headers",
				DefaultValue:     []string{},
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_ARRAY,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     3,
					},
				},
				Description: "Array of headers in Key: Value entries",
			},
			{
				Name:             "body",
				ModalDisplayName: "Body",
				DefaultValue:     "",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     4,
					},
				},
				Description: "Body contents to send in request",
			},
			{
				Name:             "socketPath",
				ModalDisplayName: "Unix Socket Path",
				CLIName:          "socketPath",
				DefaultValue:     "",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     5,
					},
				},
				Description: "Path to UNIX Socket if you want to use that instead of a remote host",
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			url, err := taskData.Args.GetStringArg("url")
			if err != nil {
				logging.LogError(err, "Failed to get url string")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			method, err := taskData.Args.GetStringArg("method")
			if err != nil {
				logging.LogError(err, "Failed to get method string")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			bodyString, err := taskData.Args.GetStringArg("body")
			if err != nil {
				logging.LogError(err, "Failed to get body string")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			taskData.Args.SetArgValue("body", base64.StdEncoding.EncodeToString([]byte(bodyString)))
			displayParams := fmt.Sprintf("%s via HTTP %s", url, method)
			socketPath, err := taskData.Args.GetStringArg("socketPath")
			if err != nil {
				logging.LogError(err, "Failed to get socketPath")
				response.Success = false
				response.Error = err.Error()
				return response
			}
			if socketPath != "" {
				displayParams += fmt.Sprintf(" to %s", socketPath)
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
