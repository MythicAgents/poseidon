package curl

import (
	// Standard
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"time"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	Url     string `json:"url"`
	Method  string `json:"method"`
	Body    string `json:"body"`
	Headers string `json:"headers"`
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
	var body []byte
	var rawHeaders []byte
	if len(args.Body) > 0 {
		body, err = base64.StdEncoding.DecodeString(args.Body)
		if err != nil {
			msg.SetError(err.Error())
			task.Job.SendResponses <- msg
			return
		}
	}
	if len(args.Headers) > 0 {
		rawHeaders, err = base64.StdEncoding.DecodeString(args.Headers)
		if err != nil {
			msg.SetError(err.Error())
			task.Job.SendResponses <- msg
			return
		}
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: tr,
	}

	var respBody []byte
	var req *http.Request
	if len(body) > 0 {
		req, err = http.NewRequest(args.Method, args.Url, bytes.NewBuffer(body))
		if err != nil {
			msg.SetError(err.Error())
			task.Job.SendResponses <- msg
			return
		}
	} else {
		req, err = http.NewRequest(args.Method, args.Url, nil)
		if err != nil {
			msg.SetError(err.Error())
			task.Job.SendResponses <- msg
			return
		}
	}
	if len(rawHeaders) > 0 {
		var headers map[string]string
		err = json.Unmarshal(rawHeaders, &headers)
		if err != nil {
			msg.SetError(err.Error())
			task.Job.SendResponses <- msg
			return
		}
		for k, v := range headers {
			if k == "Host" {
				req.Host = v
			} else {
				req.Header.Set(k, v)
			}
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	defer resp.Body.Close()
	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	msg.Completed = true
	msg.UserOutput = string(respBody)
	task.Job.SendResponses <- msg
	return
}
