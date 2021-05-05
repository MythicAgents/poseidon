# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## 1.0.0 - 2021-XX-XX

### Added

- Added support for Go modules with addition of `go.mod` file
- Added debug message output to `builder.py`
- Added build flag to omit the DWARF symbol table, `-w` from the agent to reduce file size
- Added build flag to mit the symbol table and debug information, `-s` from the agent to reduce file size
- Added empty `buildid` during agent generation
- Added `c-shared` build mode

### Changed

- Updated all Poseidon imports to use full project path
- Moved `shareedlib-darwin-linux.c` out of the project root into the `sharedlib` directory
- Moved the http and websocket C2 profiles into the profiles package and set build tags to control which is used
- Changed `builder.py` to use Go's `-X` ldflags to set C2 profile variables
- The Poseidon agent is built out of `/Mythic/agent_code` directly instead of a temp directory