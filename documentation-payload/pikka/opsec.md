+++
title = "OPSEC"
chapter = false
weight = 10
pre = "<b>1. </b>"
+++

### Post-Exploitation Jobs
All pikka commands run in a goroutine. They can't really be stopped once they start, but since most commands are quick (file ops, socks), it's usually not a big deal.

### Agent Compilation
It's a standard Go binary. If you're worried about signatures, you should probably obfuscate it yourself or use a packer. The project doesn't do any magic obfuscation out of the box.
