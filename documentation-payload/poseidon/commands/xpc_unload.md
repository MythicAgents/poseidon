+++
title = "xpc_unload"
chapter = false
weight = 132
hidden = false
+++

## Summary
Use xpc to execute routines with launchd to unload a running service.
  
- Needs Admin: False  
- Version: 1  
- Author: @xorrior, @Morpheus______, @its_a_feature_  

### Arguments

#### file

- Description: Path to the plist file to unload.
- Required Value: True  
- Default Value: None  

## Usage

```
xpc_unload -file /Users/itsafeature/Desktop/evil.plist
```


## Detailed Summary

This command uses the `xpc_pipe_routine` function to send XPC messages to `launchd`.

This uses the ROUTINE_UNLOAD routine to unload a daemon property list file.