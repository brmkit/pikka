+++
title = "objload"
chapter = false
weight = 108
hidden = false
+++

## Summary

Load and execute an object file in-memory. Windows: COFF/BOF or PE via goffloader. Linux/macOS: ELF BOF via built-in ELF loader.

* Needs Admin: False
* Version: 1
* Author: @brmk

### MITRE ATT&CK Mapping

* T1106 — Native API
* T1620 — Reflective Code Loading
* T1055 — Process Injection

## Usage

```
objload
```

## Parameters

### Default (upload a new file)

| Name        | Type      | Required | Description |
|-------------|-----------|----------|-------------|
| file_name   | File      | Yes      | Object file to load and execute (COFF/BOF, PE, or ELF BOF) |
| mode        | Choice    | Yes      | `bof` for Beacon Object Files (all platforms), `pe` for PE executables (Windows only) |
| args        | Array     | No       | Arguments to pass (see below) |
| entry_point | String    | No       | BOF entry point function name (default: `go`) |

### Existing File (reuse a previously uploaded file)

| Name          | Type      | Required | Description |
|---------------|-----------|----------|-------------|
| existing_file | Choice    | Yes      | Name of a previously uploaded file to reuse |
| mode          | Choice    | Yes      | `bof` or `pe` |
| args          | Array     | No       | Arguments to pass (see below) |
| entry_point   | String    | No       | BOF entry point function name (default: `go`) |

### BOF Arguments

BOF arguments use type-prefixed format (colon-separated):

* `z:value` — ANSI string (e.g. `z:hello`)
* `Z:value` — Wide/Unicode string (e.g. `Z:hello`)
* `i:value` — 32-bit integer (e.g. `i:1234`)
* `s:value` — 16-bit short (e.g. `s:80`)
* `b:value` — Binary data as hex (e.g. `b:4142`)

### PE Arguments

PE arguments are passed as plain strings directly to the executable.

## Detailed Summary

Unified cross-platform object file loader that replaces `coffload` with support for all major platforms. Loads and executes BOFs or PE files directly from memory without touching disk.

### Windows

Uses the [goffloader](https://github.com/praetorian-inc/goffloader) library for pure-Go in-memory execution:

* **BOF mode**: Parses the COFF object file, resolves imports (including Beacon API compatibility functions), packs typed arguments, and calls the specified entry point. Output is captured via the Beacon compatibility layer.
* **PE mode**: Loads a PE executable in-memory using the NoConsolation technique. Console output is captured and returned.

### Linux / macOS

Uses a built-in ELF loader based on [TrustedSec's ELFLoader](https://github.com/trustedsec/ELFLoader), with a CGO-based Beacon compatibility layer:

* **BOF mode only** — PE mode is not supported on non-Windows targets.
* Parses relocatable ELF objects (ET_REL), maps sections into memory with appropriate protections, and resolves relocations.
* External symbols are resolved via `dlsym` or the built-in Beacon API function table.
* Supports the standard Beacon Data and Format APIs (`BeaconDataParse`, `BeaconDataInt`, `BeaconDataExtract`, `BeaconPrintf`, `BeaconOutput`, etc.).
* x86_64 (amd64) architecture only.

### Supported OS

* Windows
* Linux
* macOS
