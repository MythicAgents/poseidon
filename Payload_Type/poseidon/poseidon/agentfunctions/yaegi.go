package agentfunctions

import (
	"errors"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

/**
Missing 1-1:

argument_class: https://github.com/github-red-tea/poseidon-redtea/blob/829a46d33f2de9bb3ee6ef1ca981e6c718904c11/Payload_Type/poseidon/mythic/agent_functions/yaegi.py#L50
	|- this doesn't seem to be used for go builds?

create_tasking: https://github.com/github-red-tea/poseidon-redtea/blob/829a46d33f2de9bb3ee6ef1ca981e6c718904c11/Payload_Type/poseidon/mythic/agent_functions/yaegi.py#L53-L67
	|- believe to have created this with TaskFunctionCreateTasking

process_response: https://github.com/github-red-tea/poseidon-redtea/blob/829a46d33f2de9bb3ee6ef1ca981e6c718904c11/Payload_Type/poseidon/mythic/agent_functions/yaegi.py#L69-L70
	|- this doesn't seem like it's needed from other cmds?
*/

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "yaegi",
		Description:         "executes a yaegi extension.",
		HelpString:          "yaegi",
		Version:             1,
		Author:              "@rookuu",
		SupportedUIFeatures: []string{"file_browser:upload"},
		MitreAttackMappings: []string{"T1620"},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:          "file_id",
				ModalDisplayName:       "File to Execute",
				Description: "File to Execute",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_FILE,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
					},
				},
			},
			{
				Name: "args",
				ModalDisplayName: "args",
				Description: "Array of arguments to pass through the program",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition: 2,
					},
				},
			},
		},

		// UNSURE IS THIS IS NEEDED?
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

		TaskFunctionCreateTasking: func(task *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			{
				response := agentstructs.PTTaskCreateTaskingMessageResponse{
					Success: true,
					TaskID:  task.Task.ID,
				}
				return response
			}
		},
	})
}
