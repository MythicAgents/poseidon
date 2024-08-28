package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

var config = agentstructs.Command{
	Name:                      "config",
	Description:               "View current config and host information",
	MitreAttackMappings:       []string{},
	TaskFunctionCreateTasking: configCreateTasking,
	Version:                   1,
}

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(config)
}

func configCreateTasking(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
	response := agentstructs.PTTaskCreateTaskingMessageResponse{
		Success: true,
		TaskID:  taskData.Task.ID,
	}
	return response
}
