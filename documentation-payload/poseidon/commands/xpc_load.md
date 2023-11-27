+++
title = "xpc_load"
chapter = false
weight = 132
hidden = false
+++

## Summary
Use xpc to load a new launch agent or launch daemon.
  
- Needs Admin: False  
- Version: 1  
- Author: @xorrior, @Morpheus______, @its_a_feature_  

### Arguments

#### file

- Description: Path to the plist file on disk to load  
- Required Value: True  
- Default Value: None

## Usage

```
xpc_load -file /Users/itsafeature/Desktop/evil.plist
```


## Detailed Summary

This command uses the `xpc_pipe_routine` function to send XPC messages to `launchd`.

This uses the ROUTINE_LOAD routine to load a daemon property list file.