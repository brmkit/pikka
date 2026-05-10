+++
title = "execute_assembly"
chapter = false
weight = 107
hidden = false
+++

## Summary

Load and execute a .NET assembly in memory on a Windows target.

* Needs Admin: False
* Version: 1
* Author: @brmk

## Usage

```
execute_assembly
```

## Detailed Summary

Loads a managed .NET assembly directly from memory and executes it without touching disk. Execution occurs within the context of the running agent process and only functions on Windows agents.

Behavior details:

* The assembly is reflectively loaded into memory using the CLR.
* By default, the `Main` method is invoked.
* Output (stdout/stderr) is captured and returned to the operator.

This command is commonly used for executing post-exploitation tooling such as credential access, situational awareness, and lateral movement frameworks entirely in memory.
