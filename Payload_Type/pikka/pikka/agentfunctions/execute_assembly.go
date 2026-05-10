package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/logging"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
	"github.com/MythicMeta/MythicContainer/utils/helpers"
)

func init() {
	agentstructs.AllPayloadData.Get("pikka").AddCommand(agentstructs.Command{
		Name:                "execute_assembly",
		HelpString:          "execute_assembly",
		Description:         "Load a .NET assembly in-memory and execute it with arguments.",
		Version:             1,
		MitreAttackMappings: []string{"T1106", "T1620", "T1105"},
		Author:              "@brmk",
		SupportedUIFeatures: []string{"file_browser:upload"},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "assembly_name",
				ModalDisplayName: "File to Upload",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_FILE,
				Description:      "Select a file to write to the remote path",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						GroupName:           "Default",
						UIModalPosition:     1,
					},
				},
			},
			{
				Name:                 "assembly_file",
				ModalDisplayName:     "Existing File",
				ParameterType:        agentstructs.COMMAND_PARAMETER_TYPE_CHOOSE_ONE,
				Description:          "Name of an existing file to use",
				DynamicQueryFunction: getFiles,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						GroupName:           "Existing File",
					},
				},
			},
			{
				Name:             "args",
				ModalDisplayName: "Assembly Arguments",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_ARRAY,
				Description:      "Arguments to pass to the assembly being executed",
				DefaultValue:     []string{},
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						GroupName:           "Default",
						UIModalPosition:     2,
					},
					{
						ParameterIsRequired: false,
						GroupName:           "Existing File",
						UIModalPosition:     2,
					},
				},
			},
		},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			return args.LoadArgsFromJSONString(input)
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return args.LoadArgsFromDictionary(input)
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}

			if taskData.Payload.OS != "Windows" {
				response.Success = false
				response.Error = "execute_assembly is only supported on Windows targets"
				return response
			}

			groupName, err := taskData.Args.GetParameterGroupName()
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}

			if groupName == "Existing File" {
				assemblyName, err := taskData.Args.GetArg("assembly_file")
				if err != nil {
					response.Success = false
					response.Error = err.Error()
					return response
				}
				search, err := mythicrpc.SendMythicRPCFileSearch(mythicrpc.MythicRPCFileSearchMessage{
					TaskID:          taskData.Task.CallbackID,
					Filename:        assemblyName.(string),
					LimitByCallback: false,
					MaxResults:      1,
				})
				if err != nil || !search.Success || len(search.Files) == 0 {
					response.Success = false
					response.Error = "Failed to find specified file"
					return response
				}

				taskData.Args.AddArg(agentstructs.CommandParameter{
					Name:         "assembly_name",
					DefaultValue: search.Files[0].AgentFileID,
					ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
						{GroupName: groupName},
					},
				})

				taskData.Args.RemoveArg("assembly_file")
				displayString := search.Files[0].Filename
				response.DisplayParams = &displayString
				return response
			}

			return response
		},
	})
}
func getFiles(input agentstructs.PTRPCDynamicQueryFunctionMessage) []string {
	fileResp, err := mythicrpc.SendMythicRPCFileSearch(mythicrpc.MythicRPCFileSearchMessage{
		LimitByCallback:     false,
		CallbackID:          input.Callback,
		IsPayload:           false,
		IsDownloadFromAgent: false,
		Filename:            "",
	})
	if err != nil {
		logging.LogError(err, "Failed to search for files in callback")
		return []string{}
	}
	if !fileResp.Success {
		logging.LogError(err, "Failed to search for files in callback", "mythic error", fileResp.Error)
		return []string{}
	}
	potentialFiles := []string{}
	for _, file := range fileResp.Files {
		if !helpers.StringSliceContains(potentialFiles, file.Filename) {
			potentialFiles = append(potentialFiles, file.Filename)
		}
	}
	return potentialFiles

}
