package taskRegistrar

import (
	"sync"

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

var commands map[string]func(structs.Task)
var commandsMapSync sync.RWMutex

func init() {
	commands = make(map[string]func(structs.Task))
}

func Register(name string, fn func(structs.Task)) {
	commandsMapSync.Lock()
	commands[name] = fn
	commandsMapSync.Unlock()
}

func Find(name string) (func(structs.Task), bool) {
	commandsMapSync.RLock()
	fn, ok := commands[name]
	commandsMapSync.RUnlock()
	return fn, ok
}
