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

Unified cross-platform object file loader with support for all major platforms. Loads and executes BOFs or PE files directly from memory without touching disk.

### Supported OS

* Windows
* Linux
* macOS

### Windows

Uses the [goffloader](https://github.com/praetorian-inc/goffloader) library for pure-Go in-memory execution:

* **BOF mode**: Parses the COFF object file, resolves imports (including Beacon API compatibility functions), packs typed arguments, and calls the specified entry point. Output is captured via the Beacon compatibility layer.
* **PE mode**: Loads a PE executable in-memory using the NoConsolation technique. Console output is captured and returned.

### Linux / macOS

Uses a built-in ELF loader derived from [TrustedSec's ELFLoader](https://github.com/trustedsec/ELFLoader), rewritten in Go with a CGo-based Beacon compatibility layer. BOF mode only - x86_64 (amd64) architecture only.

The loader works in four phases:

1. **Parsing** — Reads the ELF header, verifies it's a relocatable object (`ET_REL`) for x86_64, and parses the section headers and symbol table.

2. **Mapping** — Allocates memory with `mmap` for each `PROGBITS` section (code and initialized data) and `NOBITS` section (`.bss`). Also allocates a separate thunk table that will hold trampolines for external function calls.

3. **Relocation** — Iterates over `SHT_RELA` entries. For each external symbol (not defined in the `.o` file), the loader resolves the real address, writes it into a 12-byte trampoline (`mov rax, addr; jmp rax`), then patches the `call` instruction with the relative offset to the trampoline. For internal symbols (local functions, globals defined in the BOF), it calculates the relative offset directly.

4. **Execution** — Sets memory protections (`.text` becomes read+execute, data stays read+write), locates the entry point function in the symbol table, and invokes it through a CGo bridge.

## Symbol Resolution on Linux

On Linux a specific problem arises. In the original C implementation (ELFLoader), resolving external symbols is straightforward: `dlsym(NULL, "puts")` returns the address of `puts` because the binary is dynamically linked with libc and all its symbols are in the process's dynamic symbol table.

When the same logic runs inside a Go agent compiled with CGo, `dlsym(NULL, "puts")` returns `NULL`. Go's internal linker does not export libc symbols in the dynamic symbol table of the final binary. The libc is loaded in the process (CGo requires it), but its symbols are invisible to the standard dynamic lookup. This means any BOF that calls standard library functions () `printf`, `fopen`, `malloc`, `getuid`, essentially all of them) fails with an unresolved symbol error.

On macOS this problem does not occur: `dlsym(RTLD_DEFAULT, ...)` searches all loaded libraries regardless of how they were linked.

### Resolution strategy

The symbol resolver uses a three-level cascade:

1. **Beacon internal table** — An array maps Beacon function names (`BeaconPrintf`, `BeaconOutput`, `BeaconDataParse`, etc.) to their implementations defined in the CGo bridge. This lets BOFs call the Beacon API without modifications.

2. **Standard `dlsym`** — The loader tries `dlsym(RTLD_DEFAULT, ...)` anyway. On macOS/BSD this resolves everything. On Linux it fails for libc symbols but can resolve symbols from additional shared libraries the agent may have loaded.

3. **Compile-time symbol table (`libc_symtab.h`)** — This is the core workaround, specific to Linux. A dedicated header file contains a static array of `{name, pointer}` entries for each libc/POSIX function a BOF might call. Since the C code in the CGo block is compiled by GCC (not by Go's linker), writing `(void*)puts` produces a relocation that the dynamic linker resolves when the process starts. These pointers are always valid even though `dlsym` cannot find them. The table covers stdio, stdlib, string, ctype, POSIX (files, directories, users, processes), sockets, DNS, ifaddrs, time, mntent, and miscellaneous syscalls (roughly 160 functions).

For glibc functions implemented as macros rather than real functions (typically those from `ctype.h` like `isdigit`, `isprint`, `tolower`), the header provides wrapper functions that force a call to the function version, ensuring a valid pointer.

As a final fallback, if a symbol is not in the compile-time table, the resolver tries to open `libc.so.6`, `libm.so.6`, and `libpthread.so.0` explicitly with `RTLD_NOLOAD` (which does not load anything new, only searches already-loaded libraries) and retries `dlsym` on each handle.

### Adding new symbols

If a BOF calls a libc function not yet in the table, execution fails with `failed to resolve external symbol: <name>`. To add support:

1. Open `libc_symtab.h` in the `objload` package.
2. Add the required `#include` if not already present.
3. Add a `{"name", (void*)name}` entry in the appropriate section.
4. If the function may be a macro (`ctype.h`, `endian.h`, etc.), create a `static int _wrap_name(...)` wrapper and reference that in the table.