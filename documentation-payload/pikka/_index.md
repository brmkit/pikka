+++
title = "pikka"
chapter = false
weight = 5
+++
![logo](/agents/pikka/pikka.png?width=200px)
## Summary

**pikka** is a cross-platform (macOS, Linux, **Windows**) post-exploitation agent. It's written in Go and uses CGO for some of the OS-specific stuff.

It's a fork of poseidon, but stripped down and with some feature added. I wanted something that could run everywhere.

### Highlighted Agent Features
- Websockets for C2
- Socks5 support
- File reading
- Downloads / Uploads
- Raw execution

That's it. No fluff.

### Compilation Information
This payload uses Go to cross-compile. Since I'm using CGO, it relies on `xgo` to handle the cross-compilation for different platforms.

You have a few build options:
- `default`: Gives you a standard executable for your target OS (Windows .exe, Linux ELF, macOS Mach-O).
- `c-archive`: Creates an archive file. Useful if you want to statically link it into something else.
- `c-shared`: Creates a shared library (`.dll` for Windows, `.so` for Linux, `.dylib` for macOS).

#### c-shared

If you pick `c-shared`, you get a shared library. 
**Note:** It won't auto-execute when loaded. You need to call the exported function `RunMain`.

Test it with Python3 if you want:

```python
import ctypes
# For macOS
p = ctypes.CDLL('./pikka.dylib')
# For Linux
# p = ctypes.CDLL('./pikka.so')
# For Windows
# p = ctypes.CDLL('./pikka.dll')
p.RunMain()
```

#### c-archive

- Select `c-archive` in the build options.
- You get a zip with the `.a` file and a header.
- You can link this into your own C/C++ loader.

## Authors
- @brmk
- @xorrior (original poseidon)
- @djhohnstein (original poseidon)
- @its_a_feature_ (original poseidon)