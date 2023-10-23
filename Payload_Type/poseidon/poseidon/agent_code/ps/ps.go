package ps

import (
	// Standard
	"encoding/json"
	"regexp"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	RegexFilter string `json:"regex_filter"`
}

// Taken directly from Sliver's PS command. License file included in the folder

// Process - platform agnostic Process interface
type Process interface {
	// Pid is the process ID for this process.
	Pid() int

	// PPid is the parent process ID for this process.
	PPid() int

	// Arch is the x64 or x86
	Arch() string

	// Executable name running this process. This is not a path to the
	// executable.
	Executable() string

	// Owner is the account name of the process owner.
	Owner() string

	// bin_path of the running process
	BinPath() string

	// arguments
	ProcessArguments() []string

	//environment
	ProcessEnvironment() map[string]string

	SandboxPath() string

	ScriptingProperties() map[string]interface{}

	Name() string

	BundleID() string

	AdditionalInfo() map[string]interface{}
}

// ProcessArray - struct that will hold all of the Process results
type ProcessArray struct {
	Results []structs.ProcessDetails `json:"processes"`
}

// Run - interface method that retrieves a process list
func Run(task structs.Task) {
	procs, err := Processes()
	msg := task.NewResponse()
	params := Arguments{}
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	_ = json.Unmarshal([]byte(task.Params), &params)
	var slice []structs.ProcessDetails
	if params.RegexFilter == "" {
		// Loop over the process results and add them to the json object array
		for i := 0; i < len(procs); i++ {
			slice = append(slice, structs.ProcessDetails{
				ProcessID:             procs[i].Pid(),
				ParentProcessID:       procs[i].PPid(),
				Arch:                  procs[i].Arch(),
				User:                  procs[i].Owner(),
				BinPath:               procs[i].BinPath(),
				Arguments:             procs[i].ProcessArguments(),
				Environment:           procs[i].ProcessEnvironment(),
				SandboxPath:           procs[i].SandboxPath(),
				ScriptingProperties:   procs[i].ScriptingProperties(),
				Name:                  procs[i].Name(),
				BundleID:              procs[i].BundleID(),
				AdditionalInformation: procs[i].AdditionalInfo(),
				UpdateDeleted:         true,
			})
		}
	} else {
		for i := 0; i < len(procs); i++ {
			if exists, _ := regexp.Match(params.RegexFilter, []byte(procs[i].Name())); exists {
				slice = append(slice, structs.ProcessDetails{
					ProcessID:             procs[i].Pid(),
					ParentProcessID:       procs[i].PPid(),
					Arch:                  procs[i].Arch(),
					User:                  procs[i].Owner(),
					BinPath:               procs[i].BinPath(),
					Arguments:             procs[i].ProcessArguments(),
					Environment:           procs[i].ProcessEnvironment(),
					SandboxPath:           procs[i].SandboxPath(),
					ScriptingProperties:   procs[i].ScriptingProperties(),
					Name:                  procs[i].Name(),
					BundleID:              procs[i].BundleID(),
					AdditionalInformation: procs[i].AdditionalInfo(),
					UpdateDeleted:         false,
				})
			}
		}
	}
	jsonProcs, er := json.MarshalIndent(slice, "", "	")

	if er != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	msg.Completed = true
	msg.UserOutput = string(jsonProcs)
	msg.Processes = &slice
	task.Job.SendResponses <- msg
	return
}
