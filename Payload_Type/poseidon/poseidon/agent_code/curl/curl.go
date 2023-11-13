package curl

import (
	// Standard
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	Url         string   `json:"url"`
	Method      string   `json:"method"`
	Body        string   `json:"body"`
	Headers     []string `json:"headers"`
	SetEnv      []string `json:"setEnv"`
	ClearEnv    []string `json:"clearEnv"`
	ClearAllEnv bool     `json:"clearAllEnv"`
	GetEnv      bool     `json:"getEnv"`
}

// env are substitution environment variables to apply in the Arguments.Url and Arguments.Headers fields
var env = make(map[string]string, 0)
var envMtx = sync.RWMutex{}
var tr = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

// Run - Function that executes a curl command with Golang APIs
func Run(task structs.Task) {
	msg := task.NewResponse()
	args := &Arguments{}
	err := json.Unmarshal([]byte(task.Params), args)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	if len(args.SetEnv) > 0 {
		envMtx.Lock()
		for _, e := range args.SetEnv {
			pieces := strings.Split(e, "=")
			if len(pieces) < 2 {
				envMtx.Unlock()
				msg.SetError(fmt.Sprintf("ENVs need to be in Key=Value pairs.\nBad setting: %s\n", e))
				task.Job.SendResponses <- msg
				return
			}
			env[pieces[0]] = strings.Join(pieces[1:], "=")
		}
		envMtx.Unlock()
		msg.Completed = true
		msg.UserOutput = "Completed"
		task.Job.SendResponses <- msg
		return
	}
	if len(args.ClearEnv) > 0 {
		envMtx.Lock()
		for _, e := range args.ClearEnv {
			delete(env, e)
		}
		envMtx.Unlock()
		msg.Completed = true
		msg.UserOutput = "Removed those env settings"
		task.Job.SendResponses <- msg
		return
	}
	if args.ClearAllEnv {
		envMtx.Lock()
		env = make(map[string]string)
		envMtx.Unlock()
		msg.Completed = true
		msg.UserOutput = "Removed all env settings"
		task.Job.SendResponses <- msg
		return
	}
	if args.GetEnv {
		envMtx.Lock()
		for key, val := range env {
			msg.UserOutput += fmt.Sprintf("%s=%s\n", key, val)
		}
		envMtx.Unlock()
		msg.Completed = true
		task.Job.SendResponses <- msg
		return
	}
	var body []byte
	if len(args.Body) > 0 {
		body, err = base64.StdEncoding.DecodeString(args.Body)
		if err != nil {
			msg.SetError(err.Error())
			task.Job.SendResponses <- msg
			return
		}
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}
	url := args.Url
	envMtx.Lock()
	for key, val := range env {
		url = strings.ReplaceAll(url, fmt.Sprintf("$%s", key), val)
	}
	envMtx.Unlock()
	var respBody []byte
	var req *http.Request
	if len(body) > 0 {
		req, err = http.NewRequest(args.Method, url, bytes.NewBuffer(body))
		if err != nil {
			msg.SetError(err.Error())
			task.Job.SendResponses <- msg
			return
		}
	} else {
		req, err = http.NewRequest(args.Method, url, nil)
		if err != nil {
			msg.SetError(err.Error())
			task.Job.SendResponses <- msg
			return
		}
	}
	finalHeaders := ""
	for _, h := range args.Headers {
		headerPieces := strings.SplitN(h, ":", 2)
		envMtx.Lock()
		for key, val := range env {
			headerPieces[1] = strings.ReplaceAll(headerPieces[1], fmt.Sprintf("$%s", key), val)
		}
		envMtx.Unlock()
		finalHeaders += fmt.Sprintf("%s: %s\n", headerPieces[0], headerPieces[1])
		if headerPieces[0] == "Host" {
			req.Host = headerPieces[1]
		} else {
			req.Header.Set(headerPieces[0], headerPieces[1])
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	defer resp.Body.Close()
	initialHeaders := strings.Join(args.Headers, "\n")
	output := fmt.Sprintf("Initial URL: %s\nInitial Headers:\n%s\n", args.Url, initialHeaders)
	output += fmt.Sprintf("Final URL: %s\nFinal Headers:\n%s\nOutput:\n", url, finalHeaders)
	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		msg.UserOutput = output
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	msg.Completed = true
	msg.UserOutput = output + string(respBody)
	task.Job.SendResponses <- msg
	return
}
