+++
title = "xpc_send"
chapter = false
weight = 132
hidden = false
+++

## Summary
Use xpc to execute routines with launchd or communicate with another service/process.
  
- Needs Admin: False  
- Version: 1  
- Author: @xorrior, @Morpheus______, @its_a_feature_  

### Arguments

#### servicename

- Description: Name of the service to communicate with.
- Required Value: False  
- Default Value: None

#### data

- Description: base64 encoded json data to send to a target service  
- Required Value: False  
- Default Value: None  

## Usage

```
xpc_send -servicename com.itsafeature.test -data abc123==
```


## Detailed Summary

This command uses the `xpc_pipe_routine` function to send XPC messages to `launchd` or another service.

This sends an XPC message to the specified service endpoint.