package agentfunctions

import (
	"errors"
	"fmt"
	"github.com/MythicMeta/MythicContainer/logging"
	"path/filepath"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("poseidon").AddCommand(agentstructs.Command{
		Name:                "list_entitlements",
		Description:         "Use CSOps Syscall to list the entitlements for processes (-1 for all processes)",
		HelpString:          "list_entitlements {pid}",
		Version:             1,
		Author:              "@its_a_feature_",
		MitreAttackMappings: []string{},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{agentstructs.SUPPORTED_OS_MACOS},
		},
		AssociatedBrowserScript: &agentstructs.BrowserScript{
			ScriptPath: filepath.Join(".", "poseidon", "browserscripts", "list_entitlements_new.js"), // the name of the script in agent_browser_scripts
			Author:     "@its_a_feature_",
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:          "pid",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_NUMBER,
				DefaultValue:  -1,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     1,
					},
				},
				Description: "PID of process to query (-1 for all)",
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			if pid, err := taskData.Args.GetNumberArg("pid"); err != nil {
				logging.LogError(err, "Failed to get pid argument")
				response.Success = false
				response.Error = err.Error()
			} else if pid < 0 {
				displayString := " for all running processes"
				response.DisplayParams = &displayString
			} else {
				displayString := fmt.Sprintf(" for pid %.0f", pid)
				response.DisplayParams = &displayString
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
