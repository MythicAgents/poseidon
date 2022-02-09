+++
title = "dyld_inject"
chapter = false
weight = 120
hidden = false
+++

## Summary
Execute an application/binary and set the DYLD_INSERT_LIBRARIES environment variable to inject an arbitrary dylib on disk

- Needs Admin: False
- Version: 1
- Author: @xorrior @_r3ggi

### Arguments

#### Application

- Description: Path to the target application/binary
- Required Value: True
- Default Value: None

#### Dylibpath

- Description: Path to the dylib on disk that will be injected into the target application
- Required Value: True
- Default Value: None

#### HideApp

- Description: if set to True, the target application will be launched with the kLSLaunchAndHide flag. If set to False, the kLSLaunchDefaults flag will be used
- Required Value: True
- Default Value: False

## Usage
```
dyld_inject
```

## Detailed Summary
This function uses the LSOpenURLsWithRoles API call to launch an application/binary with the DYLD_INSERT_LIBRARIES environment variable to inject an arbitrary dylib on disk into the process
