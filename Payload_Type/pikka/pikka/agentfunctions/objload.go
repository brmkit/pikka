package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
)

func init() {
	agentstructs.AllPayloadData.Get("pikka").AddCommand(agentstructs.Command{
		Name:                "objload",
		HelpString:          "objload",
		Description:         "Load and execute an object file in-memory. Windows: COFF/BOF or PE via goffloader. Linux/macOS: ELF BOF via built-in ELF loader.",
		Version:             1,
		MitreAttackMappings: []string{"T1106", "T1620", "T1055"},
		Author:              "@brmk",
		SupportedUIFeatures: []string{},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "file_name",
				ModalDisplayName: "File to Upload",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_FILE,
				Description:      "Object file to load and execute (COFF/BOF, PE, or ELF BOF)",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						GroupName:           "Default",
						UIModalPosition:     1,
					},
				},
			},
			{
				Name:             "existing_file",
				ModalDisplayName: "Existing File",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_CHOOSE_ONE,
				Description:      "Name of an existing file to use",
				DynamicQueryFunction: getFiles,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						GroupName:           "Existing File",
						UIModalPosition:     1,
					},
				},
			},
			{
				Name:             "mode",
				ModalDisplayName: "Execution Mode",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_CHOOSE_ONE,
				Description:      "bof: execute a Beacon Object File (all platforms); pe: execute a PE in-memory (Windows only)",
				Choices:          []string{"bof", "pe"},
				DefaultValue:     "bof",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						GroupName:           "Default",
						UIModalPosition:     2,
					},
					{
						ParameterIsRequired: true,
						GroupName:           "Existing File",
						UIModalPosition:     2,
					},
				},
			},
			{
				Name:             "args",
				ModalDisplayName: "Arguments",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_ARRAY,
				Description:      "For BOF: type-prefixed args (z=string, Z=widestring, i=int, s=short, b=binary hex). For PE: plain string arguments.",
				DefaultValue:     []string{},
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						GroupName:           "Default",
						UIModalPosition:     3,
					},
					{
						ParameterIsRequired: false,
						GroupName:           "Existing File",
						UIModalPosition:     3,
					},
				},
			},
			{
				Name:             "entry_point",
				ModalDisplayName: "Entry Point",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				Description:      "BOF entry point function name (default: go)",
				DefaultValue:     "go",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						GroupName:           "Default",
						UIModalPosition:     4,
					},
					{
						ParameterIsRequired: false,
						GroupName:           "Existing File",
						UIModalPosition:     4,
					},
				},
			},
		},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{"Windows", "Linux", "macOS"},
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

			mode, err := taskData.Args.GetArg("mode")
			if err == nil {
				if modeStr, ok := mode.(string); ok && modeStr == "pe" && taskData.Payload.OS != "Windows" {
					response.Success = false
					response.Error = "PE execution mode is only supported on Windows targets"
					return response
				}
			}

			groupName, err := taskData.Args.GetParameterGroupName()
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}

			if groupName == "Existing File" {
				fileName, err := taskData.Args.GetArg("existing_file")
				if err != nil {
					response.Success = false
					response.Error = err.Error()
					return response
				}
				search, err := mythicrpc.SendMythicRPCFileSearch(mythicrpc.MythicRPCFileSearchMessage{
					TaskID:          taskData.Task.CallbackID,
					Filename:        fileName.(string),
					LimitByCallback: false,
					MaxResults:      1,
				})
				if err != nil || !search.Success || len(search.Files) == 0 {
					response.Success = false
					response.Error = "Failed to find specified file"
					return response
				}

				taskData.Args.AddArg(agentstructs.CommandParameter{
					Name:         "file_name",
					DefaultValue: search.Files[0].AgentFileID,
					ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
						{GroupName: groupName},
					},
				})

				taskData.Args.RemoveArg("existing_file")
				displayString := search.Files[0].Filename
				response.DisplayParams = &displayString
				return response
			}

			return response
		},
	})
}
