+++
title = "tsconnect"
chapter = false
weight = 130
hidden = false
+++

## Summary
Connect to a Tailscale network without a running daemon using tsnet (userspace networking). Supports exit node configuration.
  
- Needs Admin: False  
- Version: 1  
- Author: @brmk

### Arguments

#### action

- Description: Start, stop, get status, or configure as exit node for the Tailscale userspace connection.  
- Required Value: True  
- Default Value: start  
- Choices: start, stop, status, exit_node

#### auth_key

- Description: Tailscale auth key for enrollment (tskey-auth-xxx). Required for start.  
- Required Value: False (required for start action)  
- Default Value: None  

#### hostname

- Description: Hostname for the node on the tailnet.  
- Required Value: False  
- Default Value: pikka-{pid}  

#### state_dir

- Description: Directory for storing Tailscale node state. Defaults to a temp directory.  
- Required Value: False  
- Default Value: {tmpdir}/tsnet-{pid}  

#### ephemeral

- Description: Register as an ephemeral node (automatically removed when disconnected).  
- Required Value: False  
- Default Value: true  

#### control_url

- Description: Custom Tailscale control server URL (e.g. headscale). Leave empty for default.  
- Required Value: False  
- Default Value: None  

#### advertise_routes

- Description: Comma-separated CIDR prefixes to advertise as subnet routes (e.g. 10.0.0.0/24,192.168.1.0/24). Used with exit_node action.
- Required Value: False  
- Default Value: None  

#### allow_lan_access

- Description: When using exit node, allow peers to access the local LAN of this node.  
- Required Value: False  
- Default Value: false  

## Usage

```
tsconnect -action start -auth_key tskey-auth-xxxx -hostname mynode
tsconnect -action exit_node
tsconnect -action exit_node -advertise_routes 10.0.0.0/24,192.168.1.0/24 -allow_lan_access true
tsconnect -action status
tsconnect -action stop
```

## Detailed Summary
Uses the tsnet library to embed a full Tailscale node in userspace, without requiring a Tailscale daemon running on the target machine. This enables direct connectivity to a Tailscale network (tailnet) from the agent.

### Exit Node
The `exit_node` action configures the node to advertise itself as an exit node by setting routes `0.0.0.0/0` and `::/0` via the local Tailscale API (`EditPrefs`). This allows other peers on the tailnet to route all their internet traffic through the compromised host.

**Important**: After running `exit_node`, the exit node must be **approved** in the Tailscale admin console (Admin > Machines > ... > Edit route settings) or via ACL `autoApprovers` before peers can use it. With `allow_lan_access` enabled, peers using this exit node can also reach the local network of the host.

Additional subnet routes can be specified via `advertise_routes` to expose internal network segments to the tailnet.

### Workflow
1. `tsconnect -action start -auth_key tskey-auth-xxx` - Join the tailnet
2. `tsconnect -action exit_node` - Advertise as exit node  
3. Approve the exit node in Tailscale admin console
4. Peers can now select this node as their exit node
5. `tsconnect -action stop` - Leave the tailnet
