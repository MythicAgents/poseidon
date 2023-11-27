+++
title = "Development"
chapter = false
weight = 20
pre = "<b>3. </b>"
+++

## Development Environment

For command development, please use golang v1.20+

## Adding Commands

- Create a new folder with the name of the command in `Payload_Type/poseidon/poseidon/agent_code/[COMMAND]`.
- Inside the folder create a single go file called `command.go`. If the implementation of the command is compatible with both macOS and Linux, only a single go file should be necessary. 
  - If the implementation is different, create two additional files. All files that are for the darwin/macOS implementation should have the nomenclature `command_darwin.go` and `command_linux.go` for Linux. There are a minimum set of imports that are required for any command.
```
import (
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)
```

- The results/output for a command should be saved to a `Response` struct.  
- The `Completed` status should be set
- The `Status` should be set to `error` if you're erroring out of the task for some reason
- Send the resulting message out to Mythic via `task.Job.SendResponses` channel

```
func Run(task structs.Task) {
	msg := task.NewResponse()
	msg.UserOutput = "test output"
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
```
- Now that the command is created, it needs to be part of the executing switch statement. Add it to "Payload_Type/poseidon/agent_code/pkg/tasks/newTasking.go" as part of the switch statement based on the command name and be sure to import it at the top.

Please refer to the cat command in `Payload_Types/poseidon/agent_code/cat/cat.go` as an example.


## Adding C2 Profiles

- Add C2 profile code to the `Payload_Types/poseidon/pkg/profiles/` folder in file name that matches the profile 
  (e.g., `http.go` or `websocket.go`)
- Create a structure that will hold the configuration for your profile. Your C2 structure/profile should conform to 
  the Profile interface defined in `Payload_Types/poseidon/agent_code/pkg/profiles/profile.go`
- C2 profile parameters are passed in at build time using Go's `-X` ldflags options which only sets string variables
  - Create package-level _string_ variables for each C2 profile configurable option (e.g., `var USER_AGENT string`)
  - Create an `init()` function that converts the string variables to the desired format, build the profile structure, 
    and register the new struct `RegisterAvailableC2Profile(&profile)`. 
- The C2 profile must be specified as a build condition at the top of the go file (e.g., `// +build http`). 
  This ensures only the selected C2 profile is compiled into the agent
- If you're creating a new P2P profile, then make sure to also create the corresponding file in `Payload_Types/poseidon/agent_code/pkg/utils/p2p/` for your P2P profile.
  - This additional file will also have an `init` function that registers a new instance of a struct (`registerAvailableP2P(poseidonTCP{})`) that satisfies the `structs.P2PProcessor` interface.
  - This additional functionality makes it so that non-p2p agents can still connect to and communicate with your p2p profiles (ex: an `http` payload can still link to a `poseidon_tcp` even though it's not binding to a port and listening).



