package ps

import (
	// Standard
	"encoding/json"
	"regexp"
	"sync"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

var mu sync.Mutex

type Arguments struct {
	RegexFilter string `json:"regex_filter"`
}

// Taken directly from Sliver's PS command. License file included in the folder

//Process - platform agnostic Process interface
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
	ProcessEnvironment() map[string]interface{}

	SandboxPath() string

	ScriptingProperties() map[string]interface{}

	Name() string

	BundleID() string
}

//ProcessArray - struct that will hold all of the Process results
type ProcessArray struct {
	Results []structs.ProcessDetails `json:"Processes"`
}

//Run - interface method that retrieves a process list
func Run(task structs.Task) {
	procs, err := Processes()
	msg := structs.Response{}
	msg.TaskID = task.TaskID
	params := Arguments{}
	if err != nil {
		msg.SetError(err.Error())

		resp, _ := json.Marshal(msg)
		mu.Lock()
		profiles.TaskResponses = append(profiles.TaskResponses, resp)
		mu.Unlock()
		return
	}
	_ = json.Unmarshal([]byte(task.Params), &params)
	var slice []structs.ProcessDetails
	if params.RegexFilter == "" {
		// Loop over the process results and add them to the json object array
		for i := 0; i < len(procs); i++ {
			slice = append(slice, structs.ProcessDetails{
				ProcessID:           procs[i].Pid(),
				ParentProcessID:     procs[i].PPid(),
				Arch:                procs[i].Arch(),
				User:                procs[i].Owner(),
				BinPath:             procs[i].BinPath(),
				Arguments:           procs[i].ProcessArguments(),
				Environment:         procs[i].ProcessEnvironment(),
				SandboxPath:         procs[i].SandboxPath(),
				ScriptingProperties: procs[i].ScriptingProperties(),
				Name:                procs[i].Name(),
				BundleID:            procs[i].BundleID(),
			})
		}
	} else {
		for i := 0; i < len(procs); i++ {
			if exists, _ := regexp.Match(params.RegexFilter, []byte(procs[i].Name())); exists {
				slice = append(slice, structs.ProcessDetails{
					ProcessID:           procs[i].Pid(),
					ParentProcessID:     procs[i].PPid(),
					Arch:                procs[i].Arch(),
					User:                procs[i].Owner(),
					BinPath:             procs[i].BinPath(),
					Arguments:           procs[i].ProcessArguments(),
					Environment:         procs[i].ProcessEnvironment(),
					SandboxPath:         procs[i].SandboxPath(),
					ScriptingProperties: procs[i].ScriptingProperties(),
					Name:                procs[i].Name(),
					BundleID:            procs[i].BundleID(),
				})
			}
		}
	}
	jsonProcs, er := json.MarshalIndent(slice, "", "	")

	if er != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"

		resp, _ := json.Marshal(msg)
		mu.Lock()
		profiles.TaskResponses = append(profiles.TaskResponses, resp)
		mu.Unlock()
		return
	}
	msg.Completed = true
	msg.UserOutput = string(jsonProcs)
	msg.Processes = slice
	resp, _ := json.Marshal(msg)
	mu.Lock()
	profiles.TaskResponses = append(profiles.TaskResponses, resp)
	mu.Unlock()
	return
}
