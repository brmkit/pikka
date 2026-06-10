package agentfunctions

import (
	"fmt"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
)

func init() {
	agentstructs.AllPayloadData.Get("pikka").AddCommand(agentstructs.Command{
		Name:                "tsconnect",
		Description:         "Connect to a Tailscale network without a running daemon using tsnet (userspace networking).",
		HelpString:          "tsconnect -action start -auth_key tskey-auth-xxx -hostname mynode",
		Version:             1,
		Author:              "@brmk",
		MitreAttackMappings: []string{"T1572"},
		SupportedUIFeatures: []string{},
		CommandAttributes: agentstructs.CommandAttribute{
			SupportedOS: []string{},
		},
		CommandParameters: []agentstructs.CommandParameter{
			{
				Name:             "action",
				ModalDisplayName: "Action",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_CHOOSE_ONE,
				Choices:          []string{"start", "stop", "status", "exit_node"},
				DefaultValue:     "start",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: true,
						UIModalPosition:     1,
					},
				},
				Description: "Start, stop, get status, or configure as exit node for the Tailscale userspace connection",
			},
			{
				Name:             "auth_key",
				ModalDisplayName: "Auth Key",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     2,
					},
				},
				Description: "Tailscale auth key for enrollment (tskey-auth-xxx). Required for start.",
			},
			{
				Name:             "hostname",
				ModalDisplayName: "Hostname",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     3,
					},
				},
				Description: "Hostname for the node on the tailnet",
			},
			{
				Name:             "state_dir",
				ModalDisplayName: "State Directory",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     4,
					},
				},
				Description: "Directory for storing Tailscale node state. Defaults to a temp directory.",
			},
			{
				Name:             "ephemeral",
				ModalDisplayName: "Ephemeral Node",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				DefaultValue:     true,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     5,
					},
				},
				Description: "Register as an ephemeral node (automatically removed when disconnected)",
			},
			{
				Name:             "control_url",
				ModalDisplayName: "Control URL",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     6,
					},
				},
				Description: "Custom Tailscale control server URL (e.g. headscale). Leave empty for default.",
			},
			{
				Name:             "advertise_routes",
				ModalDisplayName: "Advertise Routes",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_STRING,
				DefaultValue:     "",
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     7,
					},
				},
				Description: "Comma-separated CIDR prefixes to advertise as subnet routes (e.g. 10.0.0.0/24,192.168.1.0/24)",
			},
			{
				Name:             "allow_lan_access",
				ModalDisplayName: "Allow LAN Access",
				ParameterType:    agentstructs.COMMAND_PARAMETER_TYPE_BOOLEAN,
				DefaultValue:     false,
				ParameterGroupInformation: []agentstructs.ParameterGroupInfo{
					{
						ParameterIsRequired: false,
						UIModalPosition:     8,
					},
				},
				Description: "When using exit node, allow peers to access the local LAN of this node",
			},
		},
		TaskFunctionCreateTasking: func(taskData *agentstructs.PTTaskMessageAllData) agentstructs.PTTaskCreateTaskingMessageResponse {
			response := agentstructs.PTTaskCreateTaskingMessageResponse{
				Success: true,
				TaskID:  taskData.Task.ID,
			}
			action, err := taskData.Args.GetStringArg("action")
			if err != nil {
				response.Success = false
				response.Error = err.Error()
				return response
			}
			hostname, _ := taskData.Args.GetStringArg("hostname")
			display := action
			if hostname != "" && action == "start" {
				display += fmt.Sprintf(" as %s", hostname)
			}
			if action == "exit_node" {
				display = "configure as exit node"
			}
			response.DisplayParams = &display
			return response
		},
		TaskFunctionParseArgDictionary: func(args *agentstructs.PTTaskMessageArgsData, input map[string]interface{}) error {
			return args.LoadArgsFromDictionary(input)
		},
		TaskFunctionParseArgString: func(args *agentstructs.PTTaskMessageArgsData, input string) error {
			return args.LoadArgsFromJSONString(input)
		},
	})
}
