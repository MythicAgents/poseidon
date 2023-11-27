+++
title = "xpc_submit"
chapter = false
weight = 132
hidden = false
+++

## Summary
Use xpc to execute routines with launchd to submit a keep-alive job from the specified binary without a backing plist.
  
- Needs Admin: False  
- Version: 1  
- Author: @xorrior, @Morpheus______, @its_a_feature_  

### Arguments

#### program

- Description: Path to the binary to execute.  
- Required Value: True  
- Default Value: None  

#### servicename

- Description: Name of the service to create with the specified program  
- Required Value: False  
- Default Value: None  

## Usage

```
xpc_submit -program /Users/itsafeature/Desktop/evil.bin -servicename com.itsafeature.test
```


## Detailed Summary

This command uses the `xpc_pipe_routine` function to send XPC messages to `launchd`.

This uses the ROUTINE_SUBMIT routine to submit a program for launchd to execute.