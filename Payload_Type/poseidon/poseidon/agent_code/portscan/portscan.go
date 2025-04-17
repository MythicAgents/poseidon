package portscan

import (
	// Standard
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type PortScanParams struct {
	Hosts []string `json:"hosts"` // Can also be a cidr
	Ports []string `json:"ports"`
}

func doScan(hostList []string, portListStrs []string, job *structs.Job) ([]CIDR, error) {
	// Variable declarations
	timeout := time.Duration(500) * time.Millisecond
	var portList []PortRange

	// populate the portList
	for i := 0; i < len(portListStrs); i++ {
		if strings.Contains(portListStrs[i], "-") && len(portListStrs[i]) == 1 {
			// They want all the ports
			allPorts := PortRange{1, 65535}
			var newList []PortRange
			newList = append(newList, allPorts)
			portList = newList
			break
		}
		var tmpRange PortRange
		if strings.Contains(portListStrs[i], "-") {
			parts := strings.Split(portListStrs[i], "-")
			start, err := strconv.Atoi(parts[0])
			if err == nil {
				end, err := strconv.Atoi(parts[1])
				if err == nil {
					tmpRange = PortRange{
						Start: start,
						End:   end,
					}
					portList = append(portList, tmpRange)
				}
			}
		} else {
			intPort, err := strconv.Atoi(portListStrs[i])
			if err == nil {
				tmpRange = PortRange{
					Start: intPort,
					End:   intPort,
				}
				portList = append(portList, tmpRange)
			}
		}
	}

	if len(portList) == 0 {
		err := errors.New("no ports to scan")
		return nil, err
	}

	// var cidrs []*CIDR

	var results []CIDR
	// Scan the hosts
	throttleRoutines := 10
	throttler := make(chan bool, throttleRoutines)

	for i := 0; i < len(hostList); i++ {
		newCidr, err := NewCIDR(hostList[i])
		if err != nil {
			continue
		} else {
			// Iterate through every host in hostCidr
			newCidr.ScanHosts(portList, timeout, job, throttler)
			results = append(results, *newCidr)
			// cidrs = append(cidrs, newCidr)
		}
	}
	return results, nil
}

func Run(task structs.Task) {
	msg := task.NewResponse()
	params := PortScanParams{}

	err := json.Unmarshal([]byte(task.Params), &params)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	if len(params.Hosts) == 0 || len(params.Hosts) == 1 && params.Hosts[0] == "" {
		msg.SetError("No hosts given to scan")
		task.Job.SendResponses <- msg
		return
	}
	if len(params.Ports) == 0 {
		msg.SetError("No ports given to scan")
		task.Job.SendResponses <- msg
		return
	}

	//log.Println("Beginning portscan...")
	results, err := doScan(params.Hosts, params.Ports, task.Job)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}

	// log.Println("Finished!")
	data, err := json.MarshalIndent(results, "", "    ")
	// // fmt.Println("Data:", string(data))
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	// fmt.Println("Sending on up the data:\n", string(data))
	msg.UserOutput = string(data)
	msg.Completed = true
	task.Job.SendResponses <- msg
	return
}
