package nslookup

import (
	// Standard
	"encoding/json"
	"errors"
	"net"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// важно начинать название с большой буквы - public member
type Arguments struct {
	Type    string `json:"type"`
	Address string `json:"address"`
}

// называние функции обязательно такое и с такими аргументами
func Run(task structs.Task) {
	msg := task.NewResponse()
	var args Arguments

	// тут нам падают аргументы. их нужно распарсить в структуру, которую мы ожидаем из от оператора
	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg // если не распарсилось - возвращаем ошибку на сервак
		return
	}
	address := args.Address
	reqType := args.Type

	//result of resolving
	var result any

	// resolve
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

	// send the result
	if err != nil {
		msg.SetError(err.Error())
	} else {
		data, _ := json.Marshal(result)
		msg.UserOutput = string(data)
		msg.Completed = true
	}
	task.Job.SendResponses <- msg
}
