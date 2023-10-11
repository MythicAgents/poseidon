package tasks

func Initialize() {
	go listenForNewTask()
	go listenForRemoveRunningTask()
	go listenForInboundMythicMessageFromEgressP2PChannel()
}
