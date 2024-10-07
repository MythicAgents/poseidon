package library

import (
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

var commands map[string]func(structs.Task)

func init() {
	commands = make(map[string]func(structs.Task))
}


func RegisterTask(name string, fn func(structs.Task)){
	commands[name] = fn
}


func FindTask(name string) (func(structs.Task), bool) {
	fn, ok := commands[name]
	return fn, ok
}
