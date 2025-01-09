//go:build windows
// +build windows
package ls

import (
	"os"

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

func GetPermission(finfo os.FileInfo) structs.FilePermission {
	// TODO: fixme
	return structs.FilePermission {
		UID:         0,
		GID:        0,
		Permissions: "",
		User:       "",
		Group:       "",
	}
}
