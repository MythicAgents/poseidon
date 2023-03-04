package agentfunctions

import (
	"errors"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
)

var shell = agentstructs.Command{
	Name:        "shell",
	Description: "shell yo",
	MitreAttackMappings:       []string{"T1059"},
	TaskFunctionCreateTasking: shellCreateTasking,
	TaskFunctionOPSECPost:     shellOpsecPostCheck,
	TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
		return args.LoadArgsFromDictionary(input)
	},
	Version: 1,
	TaskCompletionFunctions: map[string]agentstructs.PTTaskCompletionFunction{
		"shellCompleted":       shellCompleted,
		"pwdSubtaskCompletion": pwdSubtaskCompletion,
		"lspsGroupCompletion":  lspsGroupCompletion,
	},
}

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(shell)
}

func shellCreateTasking(taskData agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
	response := agentstructs.PTTaskCreateTaskingMessageResponse{
		Success: true,
		TaskID:  taskData.Task.ID,
	}
	//commandName := "notShell"
	//response.CommandName = &commandName
	/*
		if param1, err := taskData.Args.GetArg("param 1"); err != nil {
			logging.LogError(err, "Failed to get param 1")
		} else {
			displayParams := fmt.Sprintf("custom display params for command %s", param1)
			response.DisplayParams = &displayParams
		}
	*/
	sendMythicRPCFileCreate(taskData)
	sendMythicRPCFileSearch(taskData)
	sendMythicRPCResponseCreate(taskData)
	sendMythicRPCTaskCreateSubtask(taskData)
	sendMythicRPCTaskCreateSubtask(taskData)
	sendMythicRPCTaskCreateSubtask(taskData)
	sendMythicRPCTaskCreateSubtaskGroup(taskData)

	if _, err := mythicrpc.SendMythicRPCArtifactCreate(mythicrpc.MythicRPCArtifactCreateMessage{
		BaseArtifactType: "ProcessCreate",
		ArtifactMessage:  "/bin/sh -c " + taskData.Args.GetCommandLine(),
		TaskID:           taskData.Task.ID,
	}); err != nil {
		logging.LogError(err, "Failed to send mythicrpc artifact create")
	}
	completionFunctionName := "shellCompleted"
	response.CompletionFunctionName = &completionFunctionName
	return response
}

func shellOpsecPostCheck(task agentstructs.PTTaskMessageAllData) agentstructs.PTTaskOPSECPostTaskMessageResponse {
	response := agentstructs.PTTaskOPSECPostTaskMessageResponse{
		Success:          true,
		OpsecPostBlocked: false,
		OpsecPostMessage: "Not Blocked",
		TaskID:           task.Task.ID,
	}
	return response
}

func sendMythicRPCFileCreate(taskData agentstructs.PTTaskMessageAllData) {
	if fileResponse, err := mythicrpc.SendMythicRPCFileCreate(mythicrpc.MythicRPCFileCreateMessage{
		TaskID:       taskData.Task.ID,
		Filename:     "test",
		FileContents: []byte("this is a test"),
	}); err != nil {
		logging.LogError(err, "Failed to register new file with Mythic")
	} else if fileResponse.Success {
		logging.LogDebug("Successfully ran SendMythicRPCFileCreate")
	} else {
		logging.LogError(errors.New(fileResponse.Error), "Failed to run SendMythicRPCFileCreate")
	}
}

func sendMythicRPCFileSearch(taskData agentstructs.PTTaskMessageAllData) {
	if fileResponse, err := mythicrpc.SendMythicRPCFileSearch(mythicrpc.MythicRPCFileSearchMessage{
		CallbackID: taskData.Task.CallbackID,
		Comment:    "*",
	}); err != nil {
		logging.LogError(err, "Failed to call SendMythicRPCFileSearch")
	} else if fileResponse.Success {
		logging.LogDebug("Successfully called SendMythicRPCFileSearch")
		for _, file := range fileResponse.Files {
			// update the file
			appendBytes := []byte("appended")
			if updateResponse, err := mythicrpc.SendMythicRPCFileUpdate(mythicrpc.MythicRPCFileUpdateMessage{
				AgentFileID:    file.AgentFileId,
				Comment:        "updated comment",
				Filename:       "updated filename",
				AppendContents: &appendBytes,
			}); err != nil {
				logging.LogError(err, "Failed to call SendMythicRPCFileUpdate")
			} else if !updateResponse.Success {
				logging.LogError(errors.New(updateResponse.Error), "Failed to successfully call SendMythicRPCFileUpdate")
			} else {
				logging.LogDebug("Successfully called SendMythicRPCFileUpdate")
			}
			// get the contents
			if contentsResponse, err := mythicrpc.SendMythicRPCFileGetContent(mythicrpc.MythicRPCFileGetContentMessage{
				AgentFileID: file.AgentFileId,
			}); err != nil {
				logging.LogError(err, "Failed to call SendMythicRPCFileGetContent")
			} else if !contentsResponse.Success {
				logging.LogError(errors.New(contentsResponse.Error), "Failed to successfully run SendMythicRPCFileGetContent")
			} else {
				logging.LogDebug("Successfully called SendMythicRPCFileGetContent")
			}
		}
	} else {
		logging.LogError(errors.New(fileResponse.Error), "Failed to successfully call SendMythicRPCFileSearch")
	}
}

func sendMythicRPCResponseCreate(taskData agentstructs.PTTaskMessageAllData) {
	byteResponse := make([]byte, 256)
	for i := 0; i <= 255; i++ {
		byteResponse[i] = uint8(i)
	}
	byteResponse = append(byteResponse, byte('\n'))
	logging.LogDebug("created a byte array response for SendMythicRPCResponseCreate")
	if responseCreate, err := mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
		TaskID:   taskData.Task.ID,
		Response: byteResponse,
	}); err != nil {
		logging.LogError(err, "Failed to send sendMythicRPCResponseCreate")
	} else if !responseCreate.Success {
		logging.LogError(errors.New(responseCreate.Error), "Failed to call sendMythicRPCResponseCreate")
	} else {
		logging.LogDebug("Successfully called sendMythicRPCResponseCreate")
	}
}

func sendMythicRPCTaskCreateSubtask(taskData agentstructs.PTTaskMessageAllData) {
	subtaskFunction := "pwdSubtaskCompletion"
	if responseCreate, err := mythicrpc.SendMythicRPCTaskCreateSubtask(mythicrpc.MythicRPCTaskCreateSubtaskMessage{
		TaskID:                  taskData.Task.ID,
		CommandName:             "pwd",
		SubtaskCallbackFunction: &subtaskFunction,
	}); err != nil {
		logging.LogError(err, "Failed to send sendMythicRPCResponseCreate")
	} else if !responseCreate.Success {
		logging.LogError(errors.New(responseCreate.Error), "Failed to call sendMythicRPCResponseCreate")
	} else {
		logging.LogDebug("Successfully called sendMythicRPCResponseCreate")
	}
}
func sendMythicRPCTaskCreateSubtaskGroup(taskData agentstructs.PTTaskMessageAllData) {
	subtaskCompletionFunction := "lspsGroupCompletion"
	if responseCreate, err := mythicrpc.SendMythicRPCTaskCreateSubtaskGroup(mythicrpc.MythicRPCTaskCreateSubtaskGroupMessage{
		TaskID:                taskData.Task.ID,
		GroupCallbackFunction: &subtaskCompletionFunction,
		GroupName:             "lspsGroup Name",
		Tasks: []mythicrpc.MythicRPCTaskCreateSubtaskGroupTasks{
			{
				CommandName: "ps",
			},
			{
				CommandName: "ls",
				Params:      ".",
			},
		},
	}); err != nil {
		logging.LogError(err, "Failed to send sendMythicRPCResponseCreate")
	} else if !responseCreate.Success {
		logging.LogError(errors.New(responseCreate.Error), "Failed to call sendMythicRPCResponseCreate")
	} else {
		logging.LogDebug("Successfully called sendMythicRPCResponseCreate")
	}
}

func shellCompleted(task agentstructs.PTTaskMessageAllData, subtask *agentstructs.PTTaskMessageAllData, groupName *agentstructs.SubtaskGroupName) agentstructs.PTTaskCompletionFunctionMessageResponse {
	response := agentstructs.PTTaskCompletionFunctionMessageResponse{
		Success: true,
	}
	displayParams := "updated params"
	response.DisplayParams = &displayParams
	if responseCreate, err := mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
		TaskID:   task.Task.ID,
		Response: []byte("shellCompleted"),
	}); err != nil {
		logging.LogError(err, "Failed to send sendMythicRPCResponseCreate")
	} else if !responseCreate.Success {
		logging.LogError(errors.New(responseCreate.Error), "Failed to call sendMythicRPCResponseCreate")
	} else {
		logging.LogDebug("Successfully called sendMythicRPCResponseCreate")
	}
	return response
}
func pwdSubtaskCompletion(task agentstructs.PTTaskMessageAllData, subtask *agentstructs.PTTaskMessageAllData, groupName *agentstructs.SubtaskGroupName) agentstructs.PTTaskCompletionFunctionMessageResponse {
	response := agentstructs.PTTaskCompletionFunctionMessageResponse{
		Success: true,
	}
	if responseCreate, err := mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
		TaskID:   task.Task.ID,
		Response: []byte("\npwdSubtaskCompletion\n"),
	}); err != nil {
		logging.LogError(err, "Failed to send sendMythicRPCResponseCreate")
	} else if !responseCreate.Success {
		logging.LogError(errors.New(responseCreate.Error), "Failed to call sendMythicRPCResponseCreate")
	} else {
		logging.LogDebug("Successfully called sendMythicRPCResponseCreate")
	}
	if responseCreate, err := mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
		TaskID:   subtask.Task.ID,
		Response: []byte("\npwdSubtaskCompletion\n"),
	}); err != nil {
		logging.LogError(err, "Failed to send sendMythicRPCResponseCreate")
	} else if !responseCreate.Success {
		logging.LogError(errors.New(responseCreate.Error), "Failed to call sendMythicRPCResponseCreate")
	} else {
		logging.LogDebug("Successfully called sendMythicRPCResponseCreate")
	}
	pwdUpdatedParams := "updated from pwd"
	response.DisplayParams = &pwdUpdatedParams
	return response
}
func lspsGroupCompletion(task agentstructs.PTTaskMessageAllData, subtask *agentstructs.PTTaskMessageAllData, groupName *agentstructs.SubtaskGroupName) agentstructs.PTTaskCompletionFunctionMessageResponse {
	response := agentstructs.PTTaskCompletionFunctionMessageResponse{
		Success: true,
	}
	if responseCreate, err := mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
		TaskID:   task.Task.ID,
		Response: []byte("\nlspsGroupCompletion\n"),
	}); err != nil {
		logging.LogError(err, "Failed to send sendMythicRPCResponseCreate")
	} else if !responseCreate.Success {
		logging.LogError(errors.New(responseCreate.Error), "Failed to call sendMythicRPCResponseCreate")
	} else {
		logging.LogDebug("Successfully called sendMythicRPCResponseCreate")
	}
	if responseCreate, err := mythicrpc.SendMythicRPCResponseCreate(mythicrpc.MythicRPCResponseCreateMessage{
		TaskID:   subtask.Task.ID,
		Response: []byte("\nlspsGroupCompletion\n"),
	}); err != nil {
		logging.LogError(err, "Failed to send sendMythicRPCResponseCreate")
	} else if !responseCreate.Success {
		logging.LogError(errors.New(responseCreate.Error), "Failed to call sendMythicRPCResponseCreate")
	} else {
		logging.LogDebug("Successfully called sendMythicRPCResponseCreate")
	}
	pwdUpdatedParams := "updated from ls/ps group finishing"
	response.DisplayParams = &pwdUpdatedParams
	return response
}
