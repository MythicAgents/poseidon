+++
title = "poseidon"
chapter = false
weight = 5
+++
![logo](/agents/poseidon/poseidon.svg?width=200px)
## Summary

Poseidon is a Golang cross-platform (macOS & Linux) post-exploitation agent that leverages CGO for OS-specific API calls. 

### Highlighted Agent Features
- Websockets protocol for C2
- Socks5 in agent proxy capability
- In-memory JavaScript for Automation execution
- XPC Capability for IPC messages
- Optional HMAC+AES with EKE for encrypted comms

### Compilation Information
This payload type uses golang to cross-compile into various platforms with the help of cgo and xgo

There are three options for file types when building a poseidon payload
- The `default` option produces an executable for the selected operating system
- The `c-archive` option produces an archive file that can be used with the `sharedlib-darwin-linux.c` file to compile 
  a shared object file for Linux or dylib for macOS.
- The `c-shared` options produces a shared object file WITHOUT the ability to auto execute on load

#### c-shared

The `c-shared` build mode will return a single file, a `.so` file for Linux or a `.dylib` file for macOS. 
The shared library file does not incorporate the `sharedlib-darwin-linux.c` file.
Because of this, the library will **NOT** automatically execute when it is loaded into a process,
You must use the `c-archive` build mode for the library to automatically execute when loaded.
The returned shared library from this `c-shared` build mode requires that the `RunMain` exported function is called.

The following Python3 code can be used to test execution of the returned file:

```text
>>> import ctypes
>>> p = ctypes.CDLL('./poseidon.dylib')
>>> p.RunMain()
```

#### c-archive
- In the payload type information section of the payload creation page, please select the `c-archive` buildmode option.
- The resulting payload file will be a zip file.
  The zip contains the golang `.a` archive file, the `.h` header file and the `sharedlib-darwin-linux.c` file.
- Edit `sharedlib-darwin-linux.c` and change the `include` statement on line 7 to match the name of the golang archive 
  header file, if different.
- Execute `ranlib poseidon-darwin-10.12-amd64.a` to "update the table of contents of archive libraries".
- Use `clang` to compile a dylib on macOS:
  `clang -shared -framework Foundation -framework CoreGraphics -framework Security -framework ApplicationServices 
  -framework OSAKit -framework AppKit -fpic sharedlib-darwin-linux.c poseidon-darwin-10.12-amd64.a -o poseidon.dylib`

The following Python3 code can be used to test execution of shared library:

```text
>>> import ctypes
>>> ctypes.CDLL('./poseidon.dylib')
```

## Authors
- @xorrior
- @djhohnstein
- @its_a_feature_
