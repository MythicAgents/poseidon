+++
title = "xpc_procinfo"
chapter = false
weight = 132
hidden = false
+++

## Summary
Use xpc to execute routines with launchd to get information about a running service pid.
  
- Needs Admin: False  
- Version: 1  
- Author: @xorrior, @Morpheus______, @its_a_feature_  

### Arguments

#### pid

- Description: PID of the process  
- Required Value: False  
- Default Value: None

## Usage

```
xpc_procinfo -pid 98765
```


## Detailed Summary

This command uses the `xpc_pipe_routine` function to send XPC messages to `launchd`.

This uses the ROUTINE_DUMP_PROCESS routine to print information about the execution context of a given PID.