package agentfunctions

import (
	"strings"

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
				DefaultValue:  []string{},
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
						UIModalPosition:     3,
					},
				},
			},
			{
				Name:          "username",
				Description:   "Username for bind (Linux)",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:  "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     4,
					},
				},
			},
			{
				Name:          "password",
				Description:   "Password for bind (Linux)",
				ParameterType: agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:  "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     5,
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
				// full_path from the browser is in reversed order (DC=local,DC=example,CN=Users)
				// reverse it back to standard DN format (CN=Users,DC=example,DC=local)
				if p, ok := path.(string); ok && p != "" {
					parts := strings.Split(p, ",")
					for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
						parts[i], parts[j] = parts[j], parts[i]
					}
					input["base"] = strings.Join(parts, ",")
				}
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
