package structs

import (
	"encoding/json"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/enums/InteractiveTask"
	"os"
)

// Profile is the primary client interface for Mythic C2 profiles
// This is what starts listening/beaconing
type Profile interface {
	// ProfileName returns the name of this profile
	ProfileName() string
	// IsP2P returns if the profile is a P2P profile or not
	IsP2P() bool
	// Start is the entry point for this C2 Profile
	Start()
	// Stop the current C2 Profile
	Stop()
	// SetSleepInterval updates the sleep interval
	SetSleepInterval(interval int) string
	// SetSleepJitter updates the jitter percentage 0-100 to be used with the SleepInterval
	SetSleepJitter(jitter int) string
	// GetSleepTime returns the number of seconds to sleep before making another request using interval and jitter
	GetSleepTime() int
	// SetEncryptionKey to synchronize all c2 profiles once one has finished staging
	SetEncryptionKey(newKey string)
	// GetConfig returns a string representation of the current configuration
	GetConfig() string
	// UpdateConfig sets a parameter to a new value
	UpdateConfig(parameter string, value string)
	// GetPushChannel returns either a channel for push messages or nil
	GetPushChannel() chan MythicMessage
	// IsRunning returns if the c2 profile is currently running
	IsRunning() bool
}

// P2PProcessor is baked into the agent for all P2P profiles so that egress agents can always link to P2P profiles
type P2PProcessor interface {
	// ProfileName returns the name of this profile processor
	ProfileName() string
	ProcessIngressMessageForP2P(message *DelegateMessage)
	RemoveInternalConnection(connectionUUID string) bool
	AddInternalConnection(connection interface{})
	GetInternalP2PMap() string
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
	Action           string                    `json:"action"`
	TaskingSize      int                       `json:"tasking_size"`
	GetDelegateTasks bool                      `json:"get_delegate_tasks"`
	Delegates        *[]DelegateMessage        `json:"delegates,omitempty"`
	Responses        *[]Response               `json:"responses,omitempty"`
	Socks            *[]SocksMsg               `json:"socks,omitempty"`
	Rpfwds           *[]SocksMsg               `json:"rpfwd,omitempty"`
	Edges            *[]P2PConnectionMessage   `json:"edges,omitempty"`
	InteractiveTasks *[]InteractiveTaskMessage `json:"interactive,omitempty"`
	Alerts           *[]Alert                  `json:"alerts,omitempty"`
}

type MythicMessageResponse struct {
	Action           string                   `json:"action"`
	Tasks            []Task                   `json:"tasks"`
	Delegates        []DelegateMessage        `json:"delegates"`
	Socks            []SocksMsg               `json:"socks"`
	Rpfwds           []SocksMsg               `json:"rpfwd"`
	Responses        []json.RawMessage        `json:"responses"`
	InteractiveTasks []InteractiveTaskMessage `json:"interactive"`
}

type Task struct {
	Command           string  `json:"command"`
	Params            string  `json:"parameters"`
	Timestamp         float64 `json:"timestamp"`
	TaskID            string  `json:"id"`
	Job               *Job
	removeRunningTask chan string
}

func (t *Task) SetRemoveRunningTaskChannel(removeRunningTask chan string) {
	t.removeRunningTask = removeRunningTask
}
func (t *Task) NewResponse() Response {
	newResponse := Response{
		TaskID:            t.TaskID,
		removeRunningTask: t.removeRunningTask,
	}
	return newResponse
}

type RemoveInternalConnectionMessage struct {
	ConnectionUUID string
	C2ProfileName  string
}
type AddInternalConnectionMessage struct {
	C2ProfileName string
	Connection    interface{}
}
type InteractiveTaskMessage struct {
	TaskUUID    string                      `json:"task_id" mapstructure:"task_id"`
	Data        string                      `json:"data" mapstructure:"data"`
	MessageType InteractiveTask.MessageType `json:"message_type" mapstructure:"message_type"`
}
type Job struct {
	Stop                            *int
	ReceiveResponses                chan json.RawMessage
	SendResponses                   chan Response
	SendFileToMythic                chan SendFileToMythicStruct
	GetFileFromMythic               chan GetFileFromMythicStruct
	FileTransfers                   map[string]chan json.RawMessage
	SaveFileFunc                    func(fileUUID string, data []byte)
	RemoveSavedFile                 func(fileUUID string)
	GetSavedFile                    func(fileUUID string) []byte
	CheckIfNewInternalTCPConnection func(newInternalConnectionString string) bool
	AddInternalConnectionChannel    chan AddInternalConnectionMessage
	RemoveInternalConnectionChannel chan RemoveInternalConnectionMessage
	InteractiveTaskInputChannel     chan InteractiveTaskMessage
	InteractiveTaskOutputChannel    chan InteractiveTaskMessage
	NewAlertChannel                 chan Alert
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
	FinishedTransfer chan int
	// the following are set and used by Poseidon, Task doesn't use
	// two components used by Poseidon internally to track and handle chunk responses from Mythic for this specific file transfer
	TrackingUUID         string
	FileTransferResponse chan json.RawMessage
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
	ProcessID             int                    `json:"process_id"`
	ParentProcessID       int                    `json:"parent_process_id"`
	Arch                  string                 `json:"architecture"`
	User                  string                 `json:"user"`
	BinPath               string                 `json:"bin_path"`
	Arguments             []string               `json:"args"`
	Environment           map[string]string      `json:"env"`
	SandboxPath           string                 `json:"sandboxpath"`
	ScriptingProperties   map[string]interface{} `json:"scripting_properties"`
	Name                  string                 `json:"name"`
	BundleID              string                 `json:"bundleid"`
	UpdateDeleted         bool                   `json:"update_deleted"`
	AdditionalInformation map[string]interface{} `json:"additional_information"`
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

const (
	AlertLevelWarning string = "warning"
	AlertLevelInfo           = "info"
	AlertLevelDebug          = "debug"
)

type Alert struct {
	Source       *string                 `json:"source,omitempty"`
	Alert        string                  `json:"alert"`
	WebhookAlert *map[string]interface{} `json:"webhook_alert,omitempty"`
	Level        *string                 `json:"level,omitempty"`
	SendWebhook  bool                    `json:"send_webhook"`
}

type Response struct {
	TaskID            string               `json:"task_id"`
	UserOutput        string               `json:"user_output,omitempty"`
	Completed         bool                 `json:"completed,omitempty"`
	Status            string               `json:"status,omitempty"`
	FileBrowser       *FileBrowser         `json:"file_browser,omitempty"`
	RemovedFiles      *[]RmFiles           `json:"removed_files,omitempty"`
	Processes         *[]ProcessDetails    `json:"processes,omitempty"`
	TrackingUUID      string               `json:"tracking_uuid,omitempty"`
	Upload            *FileUploadMessage   `json:"upload,omitempty"`
	Download          *FileDownloadMessage `json:"download,omitempty"`
	Keylogs           *[]Keylog            `json:"keylogs,omitempty"`
	Artifacts         *[]Artifact          `json:"artifacts,omitempty"`
	Alerts            *[]Alert             `json:"alerts,omitempty"`
	ProcessResponse   *string              `json:"process_response,omitempty"`
	removeRunningTask chan string
}

func (r *Response) CompleteTask() {
	r.removeRunningTask <- r.TaskID
}
func (r *Response) SetError(errString string) {
	r.UserOutput = errString
	r.Status = "error"
	r.Completed = true
}

type RmFiles struct {
	Path string `json:"path"`
	Host string `json:"host"`
}
type FilePermission struct {
	UID         int    `json:"uid"`
	GID         int    `json:"gid"`
	Permissions string `json:"permissions"`
	User        string `json:"user,omitempty"`
	Group       string `json:"group,omitempty"`
}
type FileBrowser struct {
	Files         []FileData     `json:"files"`
	IsFile        bool           `json:"is_file"`
	Permissions   FilePermission `json:"permissions"`
	Filename      string         `json:"name"`
	ParentPath    string         `json:"parent_path"`
	Success       bool           `json:"success"`
	FileSize      int64          `json:"size"`
	LastModified  int64          `json:"modify_time"`
	LastAccess    int64          `json:"access_time"`
	UpdateDeleted bool           `json:"update_deleted"`
}

type FileData struct {
	IsFile       bool           `json:"is_file"`
	Permissions  FilePermission `json:"permissions"`
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
	TotalChunks int    `json:"total_chunks"`
	ChunkNum    int    `json:"chunk_num"`
	FullPath    string `json:"full_path"`
	// optionally identify a filename for the file within Mythic separate from full_path
	FileName     string `json:"filename"`
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

// SocksMsg struct for dealing with Socks and rpfwd messages
type SocksMsg struct {
	ServerId uint32 `json:"server_id"`
	Data     string `json:"data"`
	Exit     bool   `json:"exit"`
	Port     uint32 `json:"port"`
}

// Message - struct definition for websocket C2 messages
type Message struct {
	Data string `json:"data"`
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
