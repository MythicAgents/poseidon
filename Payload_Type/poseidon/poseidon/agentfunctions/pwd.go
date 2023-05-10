package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

var pwd = agentstructs.Command{
	Name:                "pwd",
	Description:         "Print the current working directory",
	Version:             1,
	MitreAttackMappings: []string{"T1083"},

	TaskFunctionOPSECPre:           pwdOpsecPreCheck,
	TaskFunctionCreateTasking:      pwdCreateTasking,
	TaskFunctionOPSECPost:          pwdOpsecPostCheck,
	TaskFunctionParseArgString:     pwdParseArgs,
	TaskFunctionParseArgDictionary: pwdParseDictArgs,
	ScriptOnlyCommand:              false,
	CommandAttributes: agentstructs.CommandAttribute{
		SupportedOS: []string{},
	},
}

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(pwd)
}

func pwdParseArgs(args *agentstructs.PTTaskMessageArgsData, input string) error {
	return nil
}

func pwdParseDictArgs(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
	return nil
}

func pwdOpsecPreCheck(task *agentstructs.PTTaskMessageAllData) agentstructs.PTTTaskOPSECPreTaskMessageResponse {
	response := agentstructs.PTTTaskOPSECPreTaskMessageResponse{
		Success:         true,
		OpsecPreBlocked: false,
		OpsecPreMessage: "Not implemented",
		TaskID:          task.Task.ID,
	}
	return response

}

func pwdCreateTasking(task *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
	response := agentstructs.PTTaskCreateTaskingMessageResponse{
		Success: true,
		TaskID:  task.Task.ID,
	}
	return response
}

func pwdOpsecPostCheck(task *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskOPSECPostTaskMessageResponse {
	response := agentstructs.PTTaskOPSECPostTaskMessageResponse{
		Success: true,
		TaskID:  task.Task.ID,
	}
	return response
}
