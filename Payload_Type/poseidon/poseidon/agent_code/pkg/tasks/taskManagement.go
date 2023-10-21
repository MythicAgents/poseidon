package tasks

import (
	"encoding/json"
	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

// listenForRemoveRunningTask uses removeRunningTasksChannel to identify tasks to remove from runningTasks
func listenForRemoveRunningTask() {
	for {
		select {
		case taskUUID := <-removeRunningTasksChannel:
			runningTaskMutex.Lock()
			delete(runningTasks, taskUUID)
			runningTaskMutex.Unlock()
		}
	}
}

// getJobListing is the 'jobs' command and prints the `ToStub` call on each task
func getJobListing(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID
	msg.Completed = true
	// For graceful error handling server-side when zero jobs are processing.
	if len(runningTasks) == 0 {
		msg.UserOutput = "0 jobs"
	} else {
		var jobList []structs.TaskStub
		runningTaskMutex.Lock()
		for _, x := range runningTasks {
			jobList = append(jobList, x.ToStub())
		}
		runningTaskMutex.Unlock()
		jsonSlices, err := json.MarshalIndent(jobList, "", "	")
		if err != nil {
			msg.UserOutput = err.Error()
			msg.Status = "error"
		} else {
			msg.UserOutput = string(jsonSlices)
		}

	}
	task.Job.SendResponses <- msg
}

// killJob is the 'jobkill' command which sets a Stop flag for the associated task to check
func killJob(task structs.Task) {
	msg := structs.Response{}
	msg.TaskID = task.TaskID

	foundTask := false
	for taskUUID, _ := range runningTasks {
		if runningTasks[taskUUID].TaskID == task.Params {
			*runningTasks[taskUUID].Job.Stop = 1
			foundTask = true
			break
		}
	}

	if foundTask {
		msg.UserOutput = fmt.Sprintf("Sent kill signal to Job ID: %s", task.Params)
		msg.Completed = true
	} else {
		msg.UserOutput = fmt.Sprintf("No job with ID: %s", task.Params)
		msg.Completed = true
	}
	task.Job.SendResponses <- msg
}
