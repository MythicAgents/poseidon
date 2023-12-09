# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## 2.0.14 - 2023-12-08

### Changed

- Fixed an issue with `shell` and `run` not reading all output from commands
  - the `.wait()` call closes the read pipes for stdout/stderr, sometimes causing output to get truncated
  
## 2.0.13 - 2023-12-08

### Changed

- Fixed the `portscan` command

## 2.0.12 - 2023-11-14

### Changed

- Updated the `ps` command to have a regular expression parameter to filter out which processes to return

## 2.0.11 - 2023-11-13

### Changed

- Updated `curl` command with `curl_env_set`, `curl_env_get`, and `curl_env_clear` commands
  - This now allows you to set "environment" variables that can be used in your `curl` command for substitutions in `url` and `headers` fields
  - The `header` field is now an array of strings formatted like `Header: Value`

## 2.0.10 - 2023-11-09

- Added `dynamichttp` v2.0 profile support

## 2.0.9 - 2023-11-03

- Added `sudo` command for macOS that leverages AuthorizationCreate and AuthorizationExecuteWithPrivileges to execute commands in an elevated context
  - You can control prompt text, prompt icon, or supply username and password to cause no prompt at all
- Updated `xpc_submit` with initial pieces of `launchctl asuser` functionality, but currently a work in progress

## 2.0.8 - 2023-11-01

- Added `clipboard` command similar to `apfell` for extraction of more than just plaintext from the clipboard

## 2.0.7 - 2023-10-30

- Updated `run` command to take KEY=VALUE environment pairs
- Removed `xpc` command
  - Added `xpc_service` command for `list`, `start`, `stop`, `enable`, `disable`, `print`, and `dumpstate` subcommands
  - Added `xpc_load`
  - Added `xpc_unload`
  - Added `xpc_submit`
  - Added `xpc_manageruid`
  - Added `xpc_send`
  - Added `xpc_procinfo`
- Updated `persist_launchd` (takes in `~` and has `remove` option)
- Updated `persist_loginitem` (supports session/global adds/removes, and adds list subcommand)
- Fixed issue in `screencapture` that would sometimes break
- Removed `dyld_inject` command since it couldn't actually start applications hidden
- Removed `execute_macho` command since it causes instability between multiple runs
  - `execute_macho` terminates the entire program when executing a universal/fat binary
- Updated `execute_memory` to `execute_library` and drop a dylib to disk and run it instead of using `NSCreateObjectFileImageFromMemory`
  - `NSCreateObjectFileImageFromMemory` writes out the macho/dylib to `/private/tmp/NSCreateObjectFileImageFromMemory-XXXX` and then calls `dlopen` on it; this removed its "memory execution" and made it much easier to detect
  - `execute_library` requires that symbols exist in the library/executable 
- Added `tail` and `head` commands
- Added `prompt` command to prompt for credentials on macOS

## 2.0.6 - 2023-10-23

- Updated `ps` command to escape JSON values in environment variables properly
- Added `test_password` functionality for macOS builds
- Added Accessibility check to `tcc_check` command
- Updated `list_entitlements` command 
- Updated `clipboard_monitor` to reduce sleep and add new lines for stdout

## 2.0.5 - 2023-10-22

- Updated agent functionality to add `tcc_check` from Justin Bui's Hermes agent

## 2.0.4 - 2023-10-20

- Updated some SOCKS processing to reduce potential for deadlocking
- Updated some PRFWD processing to reduce potential for deadlocking

## 2.0.3 - 2023-10-12

### Changed

- Fixed an issue where chunking of output messages caused messages with 0 length output to be dropped

## 2.0.2 - 2023-10-11

### Changed

- Updated the egress order to be an array instead of a dictionary for simplicity

## 2.0.1 - 2023-10-11

### Changed

- Started a big refactor of the agent design to split functions out into more isolated components

## 1.0.7 - 2023-01-10

### Changed

- Fixed an issue with the http profile leaking memory and sockets when sleep approaches 0
- Fixed an issue with the linux keylogging capabilities not reporting properly to Mythic

## 1.0.6 - 2022-8-30

### Changed

- Fixed an issue with the Linux compilation from the clipboard command

## 1.0.5 - 2022-8-26

### Added

- Added a `run` command (linux and macOS) and a `clipboard_monitor` command (macOS only) that reports data out to the user directly and to the keylogging search page.

## 1.0.4 - 2022-8-25

### Added

- Added support for ARM compilations

### Changed

- Updated P2P to leverage the new `new_uuid` JSON key instead of the `mythic_uuid` key with Mythic version 2.3.12. This change is to help reduce the number of `mythic` strings in obfuscated payloads.

## 1.0.3 - 2022-07-08

### Added

- Updated garble to not break with websockets and to support GOGARBLE=*

## 1.0.2 - 2022-05-05

### Added

- Added Garble support for Linux payloads

## 1.0.1 - 2021-06-23

### Added

- HTTP profile will use HTTP_PROXY, HTTPS_PROXY & NO_PROXY environment variables if a proxy was not explicitly provided
- Added "Ignore Proxy" to payload builder to ignore any HTTP environment variables, if any

## 1.0.0 - 2021-05-27

### Added

- Added support for Go modules with addition of `go.mod` file
- Added debug message output to `builder.py`
- Added build flag to omit the DWARF symbol table, `-w` from the agent to reduce file size
- Added build flag to mit the symbol table and debug information, `-s` from the agent to reduce file size
- Added empty `buildid` during agent generation
- Added `c-shared` build mode
- Added `execute_memory` functionality for in-proc execution of Dylibs and Mach-O executables in macOS
- Added `jsimport` functionality to load JXA scripts into memory in macOS
- Added `jsimport_call` functionality to run code from the imported scripts from `jsimport` in macOS

### Changed

- Updated all Poseidon imports to use full project path
- Moved `shareedlib-darwin-linux.c` out of the project root into the `sharedlib` directory
- Moved the http and websocket C2 profiles into the profiles package and set build tags to control which is used
- Changed `builder.py` to use Go's `-X` ldflags to set C2 profile variables
- The Poseidon agent is built out of `/Mythic/agent_code` directly instead of a temp directory
