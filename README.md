# pikka

<p align="center">
  <img alt="pikka Logo" src="documentation-payload/pikka/pikka.png" height="30%" width="30%">
</p>

So, you're looking at **pikka**.

It started as a fork of [poseidon](https://github.com/MythicAgents/poseidon), which is a cool Go agent for Mythic but only for Linux and macOS. I wanted something that could run everywhere, including Windows, but with just the essentials. So, I stripped out most of the stuff from poseidon and added just what was needed to get it running on Windows. Also, it has become my personal platform to test useful capabilities.

The goal? A minimal, multiplatform agent that I can use during testing, training and labs.

## installation

If you want to get this running in your Mythic instance, it's pretty standard. Use the `mythic-cli`:

* `sudo ./mythic-cli install github https://github.com/brmkit/pikka` (for the main branch)
* `sudo ./mythic-cli install github https://github.com/brmkit/pikka branchname` (for a specific branch)
* `sudo ./mythic-cli install folder /path/to/local/folder` (if you've already cloned it)

Once it's installed, just start it up:
`sudo ./mythic-cli start pikka`



### todo
- bof execution for windows