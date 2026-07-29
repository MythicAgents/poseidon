//go:build edit

package edit

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/enums/InteractiveTask"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

func TestReadEditableFileRejectsInvalidUTF8(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.txt")
	if err := os.WriteFile(path, []byte{0xff, 0xfe}, 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := readEditableFile(path); err == nil || !strings.Contains(err.Error(), "UTF-8") {
		t.Fatalf("expected UTF-8 validation error, got %v", err)
	}
}

func TestReplaceFileAtomicallyPreservesModeAndContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "editable.txt")
	if err := os.WriteFile(path, []byte("before"), 0640); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = replaceFileAtomically(path, []byte("after"), info.Mode()); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "after" {
		t.Fatalf("unexpected content %q", data)
	}
	updatedInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if updatedInfo.Mode().Perm() != 0640 {
		t.Fatalf("unexpected mode %o", updatedInfo.Mode().Perm())
	}
}

func TestSHA1Hex(t *testing.T) {
	if got := sha1Hex([]byte("hello")); got != "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d" {
		t.Fatalf("unexpected SHA-1 %s", got)
	}
}

func TestVerifyWrittenFileRejectsMismatchedContents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "editable.txt")
	if err := os.WriteFile(path, []byte("written contents"), 0600); err != nil {
		t.Fatal(err)
	}
	outputs := make(chan structs.InteractiveTaskMessage, 1)
	task := structs.Task{
		TaskID: "editor-task",
		Job: &structs.Job{
			InteractiveTaskOutputChannel: outputs,
		},
	}
	err := sendVerifiedSaveResult(&task, path, fileEditorRequest{
		RequestID: "save-one",
		FileID:    "staged-file",
	}, []byte("staged contents"))
	if err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("expected verification mismatch, got %v", err)
	}

	output := <-outputs
	if output.MessageType != InteractiveTask.FileEditorError {
		t.Fatalf("unexpected editor message type %d", output.MessageType)
	}
	decoded, err := base64.StdEncoding.DecodeString(output.Data)
	if err != nil {
		t.Fatal(err)
	}
	response := fileEditorResponse{}
	if err = json.Unmarshal(decoded, &response); err != nil {
		t.Fatal(err)
	}
	if response.Code != "write_verification_failed" || response.FileID != "staged-file" {
		t.Fatalf("unexpected verification failure response: %#v", response)
	}
	if response.CurrentSHA1 != sha1Hex([]byte("written contents")) {
		t.Fatalf("unexpected written SHA-1 %q", response.CurrentSHA1)
	}
}

func TestHandleSaveVerifiesAndReusesStagedFile(t *testing.T) {
	tests := []struct {
		name           string
		expectedSHA1   string
		forceOverwrite bool
	}{
		{
			name:         "ordinary save",
			expectedSHA1: sha1Hex([]byte("before")),
		},
		{
			name:           "force overwrite reuses conflict candidate",
			expectedSHA1:   "stale-sha1",
			forceOverwrite: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "editable.txt")
			if err := os.WriteFile(path, []byte("before"), 0640); err != nil {
				t.Fatal(err)
			}
			stagedData := []byte("after\n")
			getFiles := make(chan structs.GetFileFromMythicStruct, 1)
			sendFiles := make(chan structs.SendFileToMythicStruct, 1)
			outputs := make(chan structs.InteractiveTaskMessage, 1)
			responses := make(chan structs.Response, 10)
			task := structs.Task{
				TaskID: "editor-task",
				Job: &structs.Job{
					GetFileFromMythic:            getFiles,
					SendFileToMythic:             sendFiles,
					InteractiveTaskOutputChannel: outputs,
					SendResponses:                responses,
				},
			}

			go func() {
				transfer := <-getFiles
				transfer.ReceivedChunkChannel <- stagedData
				transfer.ReceivedChunkChannel <- nil
				transfer.ResultChannel <- structs.FileTransferResult{FileID: transfer.FileID}
			}()

			handleSave(&task, path, fileEditorRequest{
				Action:         "save",
				RequestID:      "save-one",
				FileID:         "staged-file",
				ExpectedSHA1:   test.expectedSHA1,
				ForceOverwrite: test.forceOverwrite,
			})

			writtenData, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(writtenData) != string(stagedData) {
				t.Fatalf("unexpected saved contents %q", writtenData)
			}

			output := <-outputs
			if output.MessageType != InteractiveTask.FileEditorResponse {
				t.Fatalf("unexpected editor message type %d", output.MessageType)
			}
			decoded, err := base64.StdEncoding.DecodeString(output.Data)
			if err != nil {
				t.Fatal(err)
			}
			response := fileEditorResponse{}
			if err = json.Unmarshal(decoded, &response); err != nil {
				t.Fatal(err)
			}
			if response.RequestID != "save-one" || response.FileID != "staged-file" {
				t.Fatalf("successful save did not reference the staged upload: %#v", response)
			}
			if response.CurrentSHA1 != sha1Hex(stagedData) {
				t.Fatalf("unexpected verified SHA-1 %q", response.CurrentSHA1)
			}
			select {
			case transfer := <-sendFiles:
				t.Fatalf("successful save unexpectedly sent a new snapshot: %#v", transfer)
			default:
			}
		})
	}
}

func TestConflictIncludesStagedFileID(t *testing.T) {
	outputs := make(chan structs.InteractiveTaskMessage, 1)
	task := structs.Task{
		TaskID: "editor-task",
		Job: &structs.Job{
			InteractiveTaskOutputChannel: outputs,
		},
	}
	sendConflict(&task, fileEditorRequest{
		RequestID:    "save-one",
		FileID:       "staged-file",
		ExpectedSHA1: "expected-sha1",
	}, "current-sha1", "conflict")

	output := <-outputs
	decoded, err := base64.StdEncoding.DecodeString(output.Data)
	if err != nil {
		t.Fatal(err)
	}
	response := fileEditorResponse{}
	if err = json.Unmarshal(decoded, &response); err != nil {
		t.Fatal(err)
	}
	if response.FileID != "staged-file" {
		t.Fatalf("conflict did not retain staged file ID: %#v", response)
	}
}

func TestCloseRequestCompletesEditorTask(t *testing.T) {
	responses := make(chan structs.Response, 1)
	task := structs.Task{
		TaskID: "editor-task",
		Job: &structs.Job{
			SendResponses: responses,
		},
	}
	requestData, err := json.Marshal(fileEditorRequest{Action: "close", RequestID: "close-one"})
	if err != nil {
		t.Fatal(err)
	}
	closed := handleRequest(&task, "", structs.InteractiveTaskMessage{
		MessageType: InteractiveTask.FileEditorRequest,
		Data:        base64.StdEncoding.EncodeToString(requestData),
	})
	if !closed {
		t.Fatal("close request should stop the editor loop")
	}
	response := <-responses
	if !response.Completed || response.Status != "completed" {
		t.Fatalf("unexpected close response: %#v", response)
	}
}
