+++
title = "Development"
chapter = false
weight = 20
pre = "<b>3. </b>"
+++

## Development Environment

For command development, please use golang v1.12+ .

## Adding Commands

- Create a new folder with the name of the command in `Payload_Types/poseidon/agent_code/[COMMAND]`.
- Inside the folder create a single go file called `command.go`. If the implementation of the command is compatible with both macOS and Linux, only a single go file should be necessary. If the implementation is different, create two additional files. All files that are for the darwin/macOS implementation should have the nomenclature `command_darwin.go` and `command_linux.go` for Linux. There are a minimum set of imports that are required for any command.
```
import (
	"pkg/utils/structs"
	"pkg/profiles"
	"encoding/json"
	"sync"
)
```
- The results/output for a command should be saved to a `Response` struct. The struct should be serialized to bytes with `json.Marshal` and then saved to the `profiles.TaskResponses` global variable. Please refer to the cat command in `Payload_Types/poseidon/agent_code/cat/cat.go` as an example.


## Adding C2 Profiles

- Add C2 profile code to the `Payload_Types/poseidon/pkg/profiles/` folder in file name that matches the profile 
  (e.g., `http.go` or `websocket.go`)
- Create a structure that will hold the configuration for your profile. Your C2 structure/profile should conform to 
  the Profile interface defined in `Payload_Types/poseidon/agent_code/pkg/profiles/profile.go`
- C2 profile parameters are passed in at build time using Go's `-X` ldflags options which only sets string variables
  - Create package-level _string_ variables for each C2 profile configurable option (e.g., `var USER_AGENT string`)
  - Create a `New()` function that converts the string variables to the desired format, build the profile structure, 
    and return a `Profile` object. The returned object should be your profile specific structure that fulfills the 
    `Profile` interface
- The C2 profile must be specified as a build condition at the top of the go file (e.g., `// +build http`). 
  This ensures only the selected C2 profile is compiled into the agent


