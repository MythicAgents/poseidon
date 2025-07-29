package structs

import (
	"encoding/json"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/enums/InteractiveTask"
	"os"
	"time"
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
	// GetSleepInterval returns the current sleep interval for the profile
	GetSleepInterval() int
	// SetSleepJitter updates the jitter percentage 0-100 to be used with the SleepInterval
	SetSleepJitter(jitter int) string
	// GetSleepJitter returns the current sleep jitter for the profile
	GetSleepJitter() int
	// GetSleepTime returns the number of seconds to sleep before making another request using interval and jitter
	GetSleepTime() int
	// Sleep performs a sleep with an optional timeout parameter
	Sleep()
	// GetKillDate returns the kill date for the profile
	GetKillDate() time.Time
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
	GetChunkSize() uint32
}

// Struct definition for CheckIn messages
type CheckInMessage struct {
	Action         string
	IPs            []string
	OS             string
	User           string
	Host           string
	Pid            int
	UUID           string
	Architecture   string
	Domain         string
	IntegrityLevel int
	ExternalIP     string
	ProcessName    string
	SleepInfo      string
	Cwd            string
}

func (e CheckInMessage) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"action":          e.Action,
		"ips":             e.IPs,
		"os":              e.OS,
		"user":            e.User,
		"host":            e.Host,
		"pid":             e.Pid,
		"uuid":            e.UUID,
		"architecture":    e.Architecture,
		"domain":          e.Domain,
		"integrity_level": e.IntegrityLevel,
		"external_ip":     e.ExternalIP,
		"process_name":    e.ProcessName,
		"sleep_info":      e.SleepInfo,
		"cwd":             e.Cwd,
	}
	return json.Marshal(alias)
}

type CheckInMessageResponse struct {
	Action string
	ID     string
	Status string
}

func (e *CheckInMessageResponse) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["action"]; ok {
		e.Action = v.(string)
	}
	if v, ok := alias["id"]; ok {
		e.ID = v.(string)
	}
	if v, ok := alias["status"]; ok {
		e.Status = v.(string)
	}
	return nil
}

// Struct definitions for EKE-RSA messages

type EkeKeyExchangeMessage struct {
	Action    string
	PubKey    string
	SessionID string
}

func (e EkeKeyExchangeMessage) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"action":     e.Action,
		"pub_key":    e.PubKey,
		"session_id": e.SessionID,
	}
	return json.Marshal(alias)
}

type EkeKeyExchangeMessageResponse struct {
	Action     string
	UUID       string
	SessionKey string
	SessionId  string
}

func (e *EkeKeyExchangeMessageResponse) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["action"]; ok {
		e.Action = v.(string)
	}
	if v, ok := alias["uuid"]; ok {
		e.UUID = v.(string)
	}
	if v, ok := alias["session_id"]; ok {
		e.SessionId = v.(string)
	}
	if v, ok := alias["session_key"]; ok {
		e.SessionKey = v.(string)
	}
	return nil
}

// Struct definitions for Tasking request messages

type MythicMessage struct {
	Action           string
	TaskingSize      int
	Delegates        *[]DelegateMessage
	Responses        *[]Response
	Socks            *[]SocksMsg
	Rpfwds           *[]SocksMsg
	Edges            *[]P2PConnectionMessage
	InteractiveTasks *[]InteractiveTaskMessage
	Alerts           *[]Alert
}

func (e MythicMessage) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"action":       e.Action,
		"tasking_size": e.TaskingSize,
	}
	if e.Delegates != nil && len(*e.Delegates) > 0 {
		alias["delegates"] = *e.Delegates
	}
	if e.Responses != nil && len(*e.Responses) > 0 {
		alias["responses"] = *e.Responses
	}
	if e.Socks != nil && len(*e.Socks) > 0 {
		alias["socks"] = *e.Socks
	}
	if e.Rpfwds != nil && len(*e.Rpfwds) > 0 {
		alias["rpfwds"] = *e.Rpfwds
	}
	if e.Edges != nil && len(*e.Edges) > 0 {
		alias["edges"] = *e.Edges
	}
	if e.InteractiveTasks != nil && len(*e.InteractiveTasks) > 0 {
		alias["interactive"] = *e.InteractiveTasks
	}
	if e.Alerts != nil && len(*e.Alerts) > 0 {
		alias["alerts"] = *e.Alerts
	}
	return json.Marshal(alias)
}

type MythicMessageResponse struct {
	Action           string
	Tasks            []Task
	Delegates        []DelegateMessage
	Socks            []SocksMsg
	Rpfwds           []SocksMsg
	Responses        []map[string]interface{}
	InteractiveTasks []InteractiveTaskMessage
}

func (e *MythicMessageResponse) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["action"]; ok {
		e.Action = v.(string)
	}
	if v, ok := alias["tasks"]; ok {
		e.Tasks = make([]Task, len(v.([]interface{})))
		for i, element := range v.([]interface{}) {
			e.Tasks[i] = Task{
				Command:   element.(map[string]interface{})["command"].(string),
				Params:    element.(map[string]interface{})["parameters"].(string),
				Timestamp: element.(map[string]interface{})["timestamp"].(float64),
				TaskID:    element.(map[string]interface{})["id"].(string),
			}
		}
	}
	if v, ok := alias["delegates"]; ok {
		e.Delegates = make([]DelegateMessage, len(v.([]interface{})))
		for i, element := range v.([]interface{}) {
			e.Delegates[i] = DelegateMessage{
				Message:       element.(map[string]interface{})["message"].(string),
				UUID:          element.(map[string]interface{})["uuid"].(string),
				C2ProfileName: element.(map[string]interface{})["c2_profile"].(string),
				MythicUUID:    element.(map[string]interface{})["new_uuid"].(string),
			}
		}
	}
	if v, ok := alias["socks"]; ok {
		e.Socks = make([]SocksMsg, len(v.([]interface{})))
		for i, element := range v.([]interface{}) {
			e.Socks[i] = SocksMsg{
				ServerId: uint32(element.(map[string]interface{})["server_id"].(float64)),
				Data:     element.(map[string]interface{})["data"].(string),
				Exit:     element.(map[string]interface{})["exit"].(bool),
				Port:     uint32(element.(map[string]interface{})["port"].(float64)),
			}
		}
	}
	if v, ok := alias["rpfwd"]; ok {
		e.Rpfwds = make([]SocksMsg, len(v.([]interface{})))
		for i, element := range v.([]interface{}) {
			e.Rpfwds[i] = SocksMsg{
				ServerId: uint32(element.(map[string]interface{})["server_id"].(float64)),
				Data:     element.(map[string]interface{})["data"].(string),
				Exit:     element.(map[string]interface{})["exit"].(bool),
				Port:     uint32(element.(map[string]interface{})["port"].(float64)),
			}
		}
	}
	if v, ok := alias["responses"]; ok {
		e.Responses = make([]map[string]interface{}, len(v.([]interface{})))
		for i, element := range v.([]interface{}) {
			e.Responses[i] = element.(map[string]interface{})
		}
	}
	if v, ok := alias["interactive"]; ok {
		e.InteractiveTasks = make([]InteractiveTaskMessage, len(v.([]interface{})))
		for i, element := range v.([]interface{}) {
			e.InteractiveTasks[i] = InteractiveTaskMessage{
				TaskUUID:    element.(map[string]interface{})["task_id"].(string),
				Data:        element.(map[string]interface{})["data"].(string),
				MessageType: InteractiveTask.MessageType(element.(map[string]interface{})["message_type"].(float64)),
			}
		}
	}
	return nil
}

type Task struct {
	Command           string
	Params            string
	Timestamp         float64
	TaskID            string
	Job               *Job
	removeRunningTask chan string
}

func (e Task) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"command":    e.Command,
		"parameters": e.Params,
		"timestamp":  e.Timestamp,
		"id":         e.TaskID,
	}
	return json.Marshal(alias)
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
	TaskUUID    string
	Data        string
	MessageType InteractiveTask.MessageType
}

func (e InteractiveTaskMessage) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"task_id":      e.TaskUUID,
		"data":         e.Data,
		"message_type": e.MessageType,
	}
	return json.Marshal(alias)
}
func (e *InteractiveTaskMessage) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["task_id"]; ok {
		e.TaskUUID = v.(string)
	}
	if v, ok := alias["data"]; ok {
		e.Data = v.(string)
	}
	if v, ok := alias["message_type"]; ok {
		e.MessageType = v.(InteractiveTask.MessageType)
	}
	return nil
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
	ProcessID             int
	ParentProcessID       int
	Arch                  string
	User                  string
	BinPath               string
	Arguments             []string
	Environment           map[string]string
	SandboxPath           string
	ScriptingProperties   map[string]interface{}
	Name                  string
	BundleID              string
	UpdateDeleted         bool
	AdditionalInformation map[string]interface{}
}

func (e ProcessDetails) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"process_id":             e.ProcessID,
		"parent_process_id":      e.ParentProcessID,
		"architecture":           e.Arch,
		"user":                   e.User,
		"bin_path":               e.BinPath,
		"args":                   e.Arguments,
		"env":                    e.Environment,
		"sandboxpath":            e.SandboxPath,
		"scripting_properties":   e.ScriptingProperties,
		"name":                   e.Name,
		"bundleid":               e.BundleID,
		"update_deleted":         e.UpdateDeleted,
		"additional_information": e.AdditionalInformation,
	}
	return json.Marshal(alias)
}

type Keylog struct {
	User        string `json:"user"`
	WindowTitle string `json:"window_title"`
	Keystrokes  string `json:"keystrokes"`
}

func (e Keylog) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"user":         e.User,
		"window_title": e.WindowTitle,
		"keystrokes":   e.Keystrokes,
	}
	return json.Marshal(alias)
}

type Artifact struct {
	BaseArtifact string `json:"base_artifact"`
	Artifact     string `json:"artifact"`
}

func (e Artifact) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"base_artifact": e.BaseArtifact,
		"Artifact":      e.Artifact,
	}
	return json.Marshal(alias)
}

const (
	AlertLevelWarning string = "warning"
	AlertLevelInfo           = "info"
	AlertLevelDebug          = "debug"
)

type Alert struct {
	Source       *string
	Alert        string
	WebhookAlert *map[string]interface{}
	Level        *string
	SendWebhook  bool
}

func (e Alert) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"alert":        e.Alert,
		"send_webhook": e.SendWebhook,
	}
	if e.Source != nil {
		alias["source"] = *e.Source
	}
	if e.Level != nil {
		alias["level"] = *e.Level
	}
	if e.WebhookAlert != nil {
		alias["webhook_alert"] = *e.WebhookAlert
	}
	return json.Marshal(alias)
}

type CallbackUpdate struct {
	Cwd                  *string
	ImpersonationContext *string
}

func (e CallbackUpdate) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{}
	if e.Cwd != nil {
		alias["cwd"] = *e.Cwd
	}
	if e.ImpersonationContext != nil {
		alias["impersonation_context"] = *e.ImpersonationContext
	}
	return json.Marshal(alias)
}

type Response struct {
	TaskID            string
	UserOutput        string
	Completed         bool
	Status            string
	FileBrowser       *FileBrowser
	RemovedFiles      *[]RmFiles
	Processes         *[]ProcessDetails
	TrackingUUID      string
	Upload            *FileUploadMessage
	Download          *FileDownloadMessage
	Keylogs           *[]Keylog
	Artifacts         *[]Artifact
	Alerts            *[]Alert
	CallbackUpdate    *CallbackUpdate
	ProcessResponse   *string
	Stdout            *string
	Stderr            *string
	removeRunningTask chan string
}

func (e Response) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"task_id":     e.TaskID,
		"user_output": e.UserOutput,
		"completed":   e.Completed,
		"status":      e.Status,
	}
	if e.FileBrowser != nil {
		alias["file_browser"] = *e.FileBrowser
	}
	if e.RemovedFiles != nil {
		alias["removed_files"] = *e.RemovedFiles
	}
	if e.Processes != nil {
		alias["processes"] = *e.Processes
	}
	if e.TrackingUUID != "" {
		alias["tracking_uuid"] = e.TrackingUUID
	}
	if e.Upload != nil {
		alias["upload"] = *e.Upload
	}
	if e.Download != nil {
		alias["download"] = *e.Download
	}
	if e.Keylogs != nil {
		alias["keylogs"] = *e.Keylogs
	}
	if e.Artifacts != nil {
		alias["artifacts"] = *e.Artifacts
	}
	if e.Alerts != nil {
		alias["alerts"] = *e.Alerts
	}
	if e.CallbackUpdate != nil {
		alias["callback"] = *e.CallbackUpdate
	}
	if e.ProcessResponse != nil {
		alias["process_response"] = *e.ProcessResponse
	}
	if e.Stdout != nil {
		alias["stdout"] = *e.Stdout
	}
	if e.Stderr != nil {
		alias["stderr"] = *e.Stderr
	}
	return json.Marshal(alias)
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
	Path string
	Host string
}

func (e RmFiles) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"path": e.Path,
		"host": e.Host,
	}
	return json.Marshal(alias)
}

type FilePermission struct {
	UID         int
	GID         int
	Permissions string
	SetUID      bool
	SetGID      bool
	Sticky      bool
	User        string
	Group       string
	Symlink     string
}

func (e FilePermission) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"uid":         e.UID,
		"gid":         e.GID,
		"permissions": e.Permissions,
		"setuid":      e.SetUID,
		"setgid":      e.SetGID,
		"sticky":      e.Sticky,
		"user":        e.User,
		"group":       e.Group,
		"symlink":     e.Symlink,
	}
	return json.Marshal(alias)
}

type FileBrowser struct {
	Files           []FileData
	IsFile          bool
	Permissions     FilePermission
	Filename        string
	ParentPath      string
	Success         bool
	FileSize        int64
	LastModified    int64
	LastAccess      int64
	UpdateDeleted   bool
	SetAsUserOutput bool
}

func (e FileBrowser) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"files":              e.Files,
		"is_file":            e.IsFile,
		"permissions":        e.Permissions,
		"name":               e.Filename,
		"parent_path":        e.ParentPath,
		"success":            e.Success,
		"size":               e.FileSize,
		"modify_time":        e.LastModified,
		"access_time":        e.LastAccess,
		"update_deleted":     e.UpdateDeleted,
		"set_as_user_output": e.SetAsUserOutput,
	}
	return json.Marshal(alias)
}

type FileData struct {
	IsFile       bool
	Permissions  FilePermission
	Name         string
	FullName     string
	FileSize     int64
	LastModified int64
	LastAccess   int64
}

func (e FileData) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"is_file":     e.IsFile,
		"permissions": e.Permissions,
		"name":        e.Name,
		"full_name":   e.FullName,
		"size":        e.FileSize,
		"modify_time": e.LastModified,
		"access_time": e.LastAccess,
	}
	return json.Marshal(alias)
}

type DelegateMessage struct {
	Message       string
	UUID          string
	C2ProfileName string
	MythicUUID    string
}

func (e DelegateMessage) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"message":    e.Message,
		"c2_profile": e.C2ProfileName,
		"new_uuid":   e.MythicUUID,
		"uuid":       e.UUID,
	}
	return json.Marshal(alias)
}
func (e *DelegateMessage) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["message"]; ok {
		e.Message = v.(string)
	}
	if v, ok := alias["uuid"]; ok {
		e.UUID = v.(string)
	}
	if v, ok := alias["c2_profile"]; ok {
		e.C2ProfileName = v.(string)
	}
	if v, ok := alias["new_uuid"]; ok {
		e.MythicUUID = v.(string)
	}
	return nil
}

type FileUploadMessage struct {
	ChunkSize   int
	TotalChunks int
	FileID      string
	ChunkNum    int
	FullPath    string
	ChunkData   string
}

func (e FileUploadMessage) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"chunk_size":   e.ChunkSize,
		"total_chunks": e.TotalChunks,
		"file_id":      e.FileID,
		"chunk_num":    e.ChunkNum,
		"full_path":    e.FullPath,
		"chunk_data":   e.ChunkData,
	}
	return json.Marshal(alias)
}

type FileDownloadMessage struct {
	TotalChunks int
	ChunkNum    int
	FullPath    string
	// optionally identify a filename for the file within Mythic separate from full_path
	FileName     string
	ChunkData    string
	FileID       string
	IsScreenshot bool
}

func (e FileDownloadMessage) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"total_chunks":  e.TotalChunks,
		"chunk_num":     e.ChunkNum,
		"full_path":     e.FullPath,
		"file_id":       e.FileID,
		"chunk_data":    e.ChunkData,
		"is_screenshot": e.IsScreenshot,
		"filename":      e.FileName,
	}
	return json.Marshal(alias)
}

type FileUploadMessageResponse struct {
	TotalChunks int
	ChunkNum    int
	ChunkData   string
	FileID      string
}

func (e *FileUploadMessageResponse) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["total_chunks"]; ok {
		e.TotalChunks = int(v.(float64))
	}
	if v, ok := alias["chunk_num"]; ok {
		e.ChunkNum = int(v.(float64))
	}
	if v, ok := alias["chunk_data"]; ok {
		e.ChunkData = v.(string)
	}
	if v, ok := alias["file_id"]; ok {
		e.FileID = v.(string)
	}
	return nil
}

type P2PConnectionMessage struct {
	Source        string
	Destination   string
	Action        string
	C2ProfileName string
}

func (e P2PConnectionMessage) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"source":      e.Source,
		"destination": e.Destination,
		"action":      e.Action,
		"c2_profile":  e.C2ProfileName,
	}
	return json.Marshal(alias)
}
func (e *P2PConnectionMessage) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["source"]; ok {
		e.Source = v.(string)
	}
	if v, ok := alias["destination"]; ok {
		e.Destination = v.(string)
	}
	if v, ok := alias["action"]; ok {
		e.Action = v.(string)
	}
	if v, ok := alias["c2_profile"]; ok {
		e.C2ProfileName = v.(string)
	}
	return nil
}

// TaskStub to post list of currently processing tasks.
type TaskStub struct {
	Command string
	Params  string
	ID      string
}

func (e TaskStub) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"command": e.Command,
		"params":  e.Params,
		"id":      e.ID,
	}
	return json.Marshal(alias)
}

type FileBrowserArguments struct {
	File        string
	Path        string
	Host        string
	FileBrowser bool
	Depth       int
}

func (e *FileBrowserArguments) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["file"]; ok {
		e.File = v.(string)
	}
	if v, ok := alias["path"]; ok {
		e.Path = v.(string)
	}
	if v, ok := alias["host"]; ok {
		e.Host = v.(string)
	}
	if v, ok := alias["file_browser"]; ok {
		e.FileBrowser = v.(bool)
	}
	if v, ok := alias["depth"]; ok {
		e.Depth = int(v.(float64))
	}
	return nil
}

// SocksMsg struct for dealing with Socks and rpfwd messages
type SocksMsg struct {
	ServerId uint32
	Data     string
	Exit     bool
	Port     uint32
}

func (e SocksMsg) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"server_id": e.ServerId,
		"data":      e.Data,
		"exit":      e.Exit,
		"port":      e.Port,
	}
	return json.Marshal(alias)
}
func (e *SocksMsg) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["server_id"]; ok {
		e.ServerId = uint32(v.(float64))
	}
	if v, ok := alias["data"]; ok {
		e.Data = v.(string)
	}
	if v, ok := alias["exit"]; ok {
		e.Exit = v.(bool)
	}
	if v, ok := alias["port"]; ok {
		e.Port = uint32(v.(float64))
	}
	return nil
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
