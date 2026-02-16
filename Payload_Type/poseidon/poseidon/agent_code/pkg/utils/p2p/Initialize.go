package p2p

func Initialize() {
	go listenForRemoveInternalP2PConnections()
	go listenForAddInternalP2PConnections()
	go listenHandleDelegateMessageForInternalP2PConnections()
}
