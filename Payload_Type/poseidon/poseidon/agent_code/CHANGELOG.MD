# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [2.2.24] - 2025-12-03

### Changed

- Updated `http` proxy usage to set both Proxy-Authorization basic header and HTTP Basic Auth headers

## [2.2.23] - 2025-12-02

### Changed

- Fixed a bug with dynamichttp configuration that would overwrite GET with POST configuration

## [2.2.22] - 2025-11-07

### Changed

- Tweaked the message processing slightly to handle more retries

## [2.2.21] - 2025-11-07

### Changed

- Adding more checks for DNS messages

## [2.2.20] - 2025-11-05

### Changed

- Updated DNS handling to be a bit more resilient and temporarily halt local ip rotation

## [2.2.19] - 2025-10-28

### Changed

- Updated MythicContainer@v1.5.2
  - Added support to supply UI position to build parameters
  
## [2.2.18] - 2025-10-07

### Changed

- Updated to MythicContainer@v1.5.0
  - Updated for dynamic parameters
- Updated to Go v1.25

## [2.2.17] - 2025-10-03

### Changed

- Fixed an issue with `httpx` usage crashing with missing parameters

## [2.2.16] - 2025-08-25

### Changed

- Updated `sleep` to take in backoff_delay and backoff_seconds fields
  - Updated logic for `sleep 0` to use backoff_delay to increase sleep time if no meaningful messages sent/received

## [2.2.15] - 2025-07-28

### Changed

- Updated the processing to minimize json tags in garble output
- Updated build process to stamp out some garble keywords in final binary

## [2.2.14] - 2025-07-21

### Changed

- Updated some processing for DNS requests to iterate over potential local addresses to find ones that work

## [2.2.13] - 2025-07-18

### Changed

- Adding initial beta support for DNS C2 

## [2.2.12] - 2025-06-18

### Changed

- Updated checkin and cd to update the new Mythic callback `cwd` fields for the UI

## [2.2.11] - 2025-06-12

### Changed

- Updated a portion of the `ls` command to return 0 instead of -1 on error

## [2.2.10] - 2025-06-05

### Changed

- Updated the way Poseidon handles callback host/callback port and proxy host/proxy port to be the same

## 2.2.9 - 2025-06-02

### Changed

- Merged PR for updated sshauth command
- Updated sshauth command to use Credential type
- Updated sshauth agent code for rate limiting to 10 attemps at a time
- Updated CIDR parsing code to respect /32
- Updated ls to add `d` identifiers in permissions listings

## 2.2.8 - 2025-05-30

### Changed

- Updated the download browserscript to render as the download happens
- Updated the download task to not report back periodic progress updates
- Added the supported ui feature `task:job_kill` to `jobkill` for support through the UI

## 2.2.7 - 2025-05-19

### Changed

- Added some additional checks in the websocket code for failed connections
- Added some http idle connection clearing code in other c2 profiles

## 2.2.6 - 2025-04-17

### Changed

- Updated the portscan command to throttle to 10 hosts at a time and respect kill commands quicker

## 2.2.5 - 2025-03-19

### Changed

- Updated `ls` to not mark files as deleted if it runs into errors listing files

## 2.2.4 - 2025-03-12

### Changed

- Updated MythicContainer version to v1.4.19
- Added default value for lsopen args

## 2.2.3 - 2025-03-05

### Changed

- Updated MythicContainer version to v1.4.18
- Updated download browser script to show current download percentage

## 2.2.2 - 2025-02-24

### Changed

- Updated `upload` to allow specifying existing filename to upload
- Updated `ls` to report back symlink data to browser script outputs
- Added `chmod` command
- Updated `rm` display parameters
- Updated `ls` browser script to support new features
- Updated `ls` to allow a `depth` parameter for recursively doing file listings
- Fixed a bug with `persist_launchd` if no path was specified

## 2.2.0 - 2025-02-10

### Changed

- Updated `ls` to reflect SUID/SGID/Sticky bits in permissions string
- Removed `poseidon_tcp`
- Added `tcp` support for ability to link with `apollo` via TCP as well

## 2.1.15 - 2025-01-17

### Changed

- Updated the `ls` command to not double return output and leverage the new SetAsUserOutput flag in file browser data
- Updated the `socks` command to continue on UDP timeouts and accept responses from different source UDP ports

## 2.1.14 - 2024-12-08

### Changed

- Updated the httpx rotation method for random in success as well as error

## 2.1.13 - 2024-11-27

### Changed

- Updated the egress order to include httpx

## 2.1.12 - 2024-11-26 

### Changed

- Updated httpx implementation to match latest v0.0.0.15 release of httpx

## 2.1.11 - 2024-11-15

### Changed

- Refactored the egress profiles to use a custom .Sleep() function that can be interrupted
  - this allows changing sleep times to take effect immediately instead of waiting another sleep interval
- Updated the checkAlive function to be more accurate

## 2.1.10 - 2024-11-13

### Changed

- Merged in PR to add `lsopen` command
- Merged in PR to update `portscan` for handling IPv6 addresses and updating browser script

## 2.1.9 - 2024-10-24

- Fixed issue with new UDP associate breaking for non-udp long-lived connections

## 2.1.8 - 2024-10-14

### Changed

- Updated `socks` to support UDP Associate
- Updated `socks` to support Username/Password Auth for SOCKS connections

## 2.1.6 - 2024-09-06

### Changed

- Updated jobs/jobkill to track properly
- Updated ssh and sshauth to kill connections 

## 2.1.3 - 2024-09-05

### Changed

- Added ifconfig command
- Added caffeinate command for macOS
- Added ssh interactive command similar to pty, but for ssh connections

## 2.1.2 - 2024-08-05

### Changed

- Adjusted the args parsing for `run` to not have env get auto populated
- Adjusted the `pty` command to not automatically spawn an extra port on the Mythic server
- Adjusted the `sudo` command to allow for parameters
- Added new `shell_config` command to allow changing the shell used for `shell` commands
- Added `config` command to view basic host config again dynamically pulled
## 2.1.2 - 2024-07-11 

### Changed

- fixed an issue with sshauth that didn't properly validate private keys before use

## 2.1.1 - 2024-07-11

### Changed

- fixed another issue with `ps` on some macOS machines that would cause issues with the RDProcess library used
## 2.0.36 - 2024-04-24

### Changed

- Fixed an issue with `ps` on new macOS machines that would cause the plist parsing to crash Poseidon

## 2.0.35 - 2024-04-11

### Changed

- Fixed an issue with poseidon_tcp values not getting stamped in correctly during compile time due to bad variable name

## 2.0.34 - 2024-04-09

### Changed

- Fixed an issue where dynamichttp profiles weren't getting serialized properly for configuration

## 2.0.33 - 2024-03-29

### Changed

- fix reverse append function

## 2.0.32 - 2024-03-29

### Changed

- Added initial beta support for the `httpx` profile

## 2.0.31 - 2024-03-27

### Changed

- The websocket profile treats all values as strings instead of numbers and strings
  - added extra checks to catch conversion errors and convert them properly before stamping into agent
  
## 2.0.30 - 2024-03-19

### Changed

- Updated the dylib/so build process (c-shared mode) to not require additional compilation
  - This leverages a build flow from [merlin](https://github.com/MythicAgents/merlin/) to conditionally add in constructors to auto run on load instead of waiting for RunMain execution
  
## 2.0.29 - 2024-03-14

### Changed

- Fixed the build process that had a bad type used

## 2.0.28 - 2024-03-13

### Changed

- Fixed an issue with `listtasks` causes a segfault when not root
- Updated the build process to take in base64 encoded values to help reduce easily visible data from build

## 2.0.27 - 2024-03-08

### Changed

- `getenv` now sorts the environment variables

## 2.0.26 - 2024-03-08

### Changed

- The latest ATT&CK removed `T1547.011` and replaced it with `T1647` which led to warnings in Mythic's UI

## 2.0.25 - 2024-03-06

### Changed

- Fixed the `ls` command to handle surrounding quotes
- Added `static` option to statically compile poseidon

## 2.0.24 - 2024-02-29

### Changed

- Fixed the `cd` command to handle surrounding quotes
- Updated a few other commands with default values that were missing them
- Updated go.mod files

## 2.0.22 - 2024-02-15

### Changed

- Fixed an issue with prompt command not releasing the prompt window on successful input
- Fixed an issue with the clipboard_monitor crashing when it couldn't find the window name
- Fixed an issue with the sudo command on macOS not respecting and changing the new process's UID/GID

## 2.0.20 - 2024-01-10

### Changed

- Updated the http response processing to always close the body since it might prevent additional connections

## 2.0.19 - 2023-12-26

### Changed

- Fixed an issue with websocket profile during egress rotation

## 2.0.18 - 2023-12-22

### Changed

- Updated the dynamichttp profile to check for len 0 Get or Post options

## 2.0.17 - 2023-12-21

### Changed

- Updated rpfwd command to track multiple listening ports and report back the `port` attribute with rpfwd data

## 2.0.16 - 2023-12-21

### Changed

- Change errors into debug prints for mismatched server values in headers for dynamichttp profile

## 2.0.15 - 2023-12-21

### Changed

- Fixed the name of the agent configuration variable in the dynamichttp profile to match the one from the profile itself

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
