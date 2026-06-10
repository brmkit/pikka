# pikka

<p align="center">
  <img alt="pikka Logo" src="documentation-payload/pikka/pikka.png" height="30%" width="30%">
</p>

So, you're looking at **pikka**.

It started as a fork of [poseidon](https://github.com/MythicAgents/poseidon), which is a cool Go agent for Mythic but only for Linux and macOS. I wanted something that could run everywhere, including Windows, but with just the essentials. So, I stripped out most of the stuff from poseidon and added just what was needed to get it running on Windows. Also, it has become my personal platform to test useful capabilities.

The goal? A minimal, multiplatform agent that I can use during testing, training and labs.

### installation

If you want to get this running in your Mythic instance, it's pretty standard. Use the `mythic-cli`:

* `sudo ./mythic-cli install github https://github.com/brmkit/pikka` (for the main branch)
* `sudo ./mythic-cli install github https://github.com/brmkit/pikka branchname` (for a specific branch)
* `sudo ./mythic-cli install folder /path/to/local/folder` (if you've already cloned it)

Once it's installed, just start it up:
`sudo ./mythic-cli start pikka`

### capabilities

Additional capabilities on top of the base poseidon command set, tested on Linux and Windows:

- **tsconnect**: embeds a Tailscale node in userspace via tsnet, joining a tailnet without a running daemon. Supports ephemeral nodes, custom control servers (headscale), exit node advertisement, and subnet route advertising.
- **objload**: in-memory object file loader and executor. On Windows, loads COFF/BOF and PE files via goffloader. On Linux/macOS, loads ELF BOF files via a built-in ELF loader with Beacon API compatibility layer (based on TrustedSec's ELFLoader).
- **execute_assembly**: reflectively loads and executes .NET assemblies in-memory using the CLR. Windows only.
- **exit**: terminates the agent with self-deletion. On Windows, renames the binary to an ADS and deletes it. On Linux/macOS, forks a cleanup process before exiting.
- **ldapsearch**: executes LDAP queries against a directory service and returns results in raw and structured formats. Supports custom base DN, filters, and attribute selection.

### disclaimer

Yes, this repository is developed with AI support. Sometimes with local models for minimal tasks (testing usage patterns and optimizations, minor fix), sometimes with frontier models for heavier lifting (code generation, documentation, debugging). The design, direction, and review are mine, the AI is just a tool in the process.

I'm not looking for anyone's approval on this. It's just honest to say it.