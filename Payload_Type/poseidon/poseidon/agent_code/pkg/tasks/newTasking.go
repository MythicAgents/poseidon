package tasks

import (
	"os"

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/tasks/taskRegistrar"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

var newTaskChannel = make(chan structs.Task, 10)

func init() {
	taskRegistrar.Register("exit", func(_ structs.Task) {
		os.Exit(0)
	})
}

// listenForNewTask uses NewTaskChannel to spawn goroutine based on task's Run method
func listenForNewTask() {
	for {
		task := <-newTaskChannel
		f, ok := taskRegistrar.Find(task.Command)
		if ok {
			go f(task)
		} else {
			resp := task.NewResponse()
			resp.SetError("Unknown command")
			task.Job.SendResponses <- resp
		}
	}
}
