package agentfunctions

import (
	"errors"
	"path/filepath"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "portscan",
		Description:         "Scan host(s) for open ports.",
		HelpString:          "portscan",
		Version:             1,
		Author:              "@djhohnstein",
		MitreAttackMappings: []string{"T1046"},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		AssociatedBrowserScript: &agentstructs.BrowserScript{
			ScriptPath: filepath.Join(".", "poseidon", "browserscripts", "portscan_new.js"),
			Author:     "@djhohnstein",
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "hosts",
				ModalDisplayName: "Hosts to scan",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_ARRAY,
				DefaultValue:     []string{},
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
					},
				},
				Description: "List of host IPs or CIDR notations",
			},
			{
				Name:          "ports",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_ARRAY,
				DefaultValue:  []string{},
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     2,
					},
				},
				Description: "List of ports to scan. Can use the dash separator to specify a range.",
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
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
