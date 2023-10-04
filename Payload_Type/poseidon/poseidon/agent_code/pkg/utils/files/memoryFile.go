package files

import "sync"

// Map used to handle go routines that are waiting for a response from apfell to continue
var storedFiles = make(map[string][]byte)
var storedFilesMutex sync.RWMutex

func SaveToMemory(fileUUID string, data []byte) {
	storedFilesMutex.Lock()
	defer storedFilesMutex.Unlock()
	storedFiles[fileUUID] = data
}

func RemoveFromMemory(fileUUID string) {
	storedFilesMutex.Lock()
	defer storedFilesMutex.Unlock()
	delete(storedFiles, fileUUID)
}

func GetFromMemory(fileUUID string) []byte {
	storedFilesMutex.RLock()
	defer storedFilesMutex.RUnlock()
	if data, ok := storedFiles[fileUUID]; ok {
		return data
	} else {
		return nil
	}
}
