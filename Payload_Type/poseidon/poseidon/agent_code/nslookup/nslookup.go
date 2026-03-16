package nslookup

import (
	"encoding/json"
	"errors"
	"net"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	Type    string `json:"type"`
	Address string `json:"address"`
}

func Run(task structs.Task) {
	msg := task.NewResponse()
	var args Arguments

	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg 
		return
	}
	address := args.Address
	reqType := args.Type

	var result any

	if reqType == "A" {
		result, err = net.LookupHost(address)
	} else if reqType == "PTR" {
		result, err = net.LookupAddr(address)
	} else if reqType == "TXT" {
		result, err = net.LookupTXT(address)
	} else if reqType == "MX" {
		result, err = net.LookupMX(address)
	} else if reqType == "CNAME" {
		result, err = net.LookupCNAME(address)
	} else if reqType == "NS" {
		result, err = net.LookupNS(address)
	} else {
		err = errors.New("invalid request type")
	}

	if err != nil {
		msg.SetError(err.Error())
	} else {
		data, _ := json.Marshal(result)
		msg.UserOutput = string(data)
		msg.Completed = true
	}
	task.Job.SendResponses <- msg
}
