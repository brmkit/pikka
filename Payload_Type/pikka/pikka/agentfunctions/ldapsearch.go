package agentfunctions

import (
	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("pikka").AddCommand(agentstructs.Command{
		Name:                "ldapsearch",
		Description:         "A simple LDAP search command",
		Version:             2,
		Author:              "@brmk",
		SupportedUIFeatures: []string{"ldapsearch", "ldap_browser:view", "ldap_browser:list"},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:          "query",
				Description:   "LDAP query string",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:  "(&(objectclass=top)(objectclass=container))",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     0,
					},
				},
			},
			{
				Name:          "base",
				Description:   "Base DN for the search",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:  "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						// TODO
						ParameterIsRequired: false,
						UIModalPosition:     1,
					},
				},
			},
			{
				Name:          "attributes",
				Description:   "List of attributes to return",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_ARRAY,
				DefaultValue:  []string{""},
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     2,
					},
				},
			},
			{
				Name:          "limit",
				Description:   "Limit results",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_NUMBER,
				DefaultValue:  0,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     4,
					},
				},
			},
			{
				Name:          "username",
				Description:   "Username for bind (Linux)",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:  "not used in Windows",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     5,
					},
				},
			},
			{
				Name:          "password",
				Description:   "Password for bind (Linux)",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:  "not used in Windows",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     6,
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

			if path, ok := input["full_path"]; ok {
				input["base"] = path
			}

			return args.LoadArgsFromDictionary(input)
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}

			// if base == callbackhost than ""
			base, _ := taskData.Args.GetArg("base")
			if base == taskData.Callback.Host {
				taskData.Args.SetArgValue("base", "")
			}

			return response
		},
		TaskFunctionProcessResponse: func(processResponse agentstructs.PtTaskProcessResponseMessage) agentstructs.PTTaskProcessResponseMessageResponse {
			response := agentstructs.PTTaskProcessResponseMessageResponse{
				TaskID:  processResponse.TaskData.Task.ID,
				Success: true,
			}
			return response
		},
	})
}
