package tasks

// listenForRemoveRunningTask uses removeRunningTasksChannel to identify tasks to remove from runningTasks
func listenForRemoveRunningTask() {
	for taskUUID := range removeRunningTasksChannel {
		runningTaskMutex.Lock()
		delete(runningTasks, taskUUID)
		runningTaskMutex.Unlock()
	}
}
