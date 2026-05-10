+++
title = "ldapsearch"
chapter = false
weight = 107
hidden = false
+++

## Summary

Execute LDAP search queries against a directory service and return the results in both raw and structured (browser-friendly) formats.

* Needs Admin: False
* Version: 1
* Author: @brmk

## Usage

```
ldapsearch
```

## Detailed Summary
The ldapsearch command allows an operator to perform LDAP queries directly from the agent. The behavior varies slightly depending on the operating system, but the goal is the same: enumerate AD objects and attributes using standard LDAP query.

#### windows
- LDAP queries are executed using the context of the running process.
- No explicit credentials are required.

#### linux / macOS (Darwin)
- LDAP queries are executed using explicitly provided credentials.


Use this command with [ldap_browser](https://github.com/MythicC2Profiles/ldap_browser/).