package agentfunctions

import (
	"errors"
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "kill",
		Description:         "Kill a process by specifying a PID",
		HelpString:          "kill [pid]",
		Version:             1,
		Author:              "@xorrior",
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			return response
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			if len(args.GetCommandLine()) == 0 {
				return errors.New("must supply a PID")
			} else {
				return nil
			}
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			return nil
		},
	})
}
