+++
title = "xpc_manageruid"
chapter = false
weight = 132
hidden = false
+++

## Summary
Use xpc to execute routines with launchd to determine the UID of the current user context
  
- Needs Admin: False  
- Version: 1  
- Author: @xorrior, @Morpheus______, @its_a_feature_  

### Arguments

## Usage

```
xpc_manageruid
```


## Detailed Summary

This command uses the `xpc_pipe_routine` function to send XPC messages to `launchd`.