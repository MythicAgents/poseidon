package main

import (
	poseidonfunctions "MyContainer/poseidon/agentfunctions"
	poseidontcpfunctions "MyContainer/poseidon_tcp/c2functions"
	"github.com/MythicMeta/MythicContainer"
)

func main() {
	// load up the agent functions directory so all the init() functions execute
	poseidonfunctions.Initialize()
	poseidontcpfunctions.Initialize()
	// sync over definitions and listen
	MythicContainer.StartAndRunForever([]MythicContainer.MythicServices{
		MythicContainer.MythicServicePayload,
		MythicContainer.MythicServiceC2,
	})
}
