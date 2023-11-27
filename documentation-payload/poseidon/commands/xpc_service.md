+++
title = "xpc_service"
chapter = false
weight = 132
hidden = false
+++

## Summary
Use xpc to execute routines with launchd.
  
- Needs Admin: False  
- Version: 1  
- Author: @xorrior, @Morpheus______, @its_a_feature_  

### Parameter Groups

#### Print - Arguments

##### servicename
- Description: The name of the service to print. Leaving this empty will print all services
- Required Value: False
- Default Value: None

##### print
- Description: Boolean flag to indicate if you want to print information about the specified service or all services
- Required Value: False
- Default Value: False

#### list - Arguments

##### servicename
- Description: The name of the service to list. Leaving this empty will list all services
- Required Value: False
- Default Value: None

##### list
- Description: Boolean flag to indicate if you want to list information about the specified service or all services
- Required Value: False
- Default Value: False

#### start - Arguments

##### servicename
- Description: The name of the service to start.
- Required Value: True
- Default Value: None

##### start
- Description: Boolean flag to indicate if you want to start a specific service
- Required Value: False
- Default Value: False

#### stop - Arguments

##### servicename
- Description: The name of the service to stop.
- Required Value: True
- Default Value: None

##### stop
- Description: Boolean flag to indicate if you want to stop a specific service
- Required Value: False
- Default Value: False

#### disable - Arguments

##### servicename
- Description: The name of the service to disable.
- Required Value: True
- Default Value: None

##### disable
- Description: Boolean flag to indicate if you want to disable a specific service
- Required Value: False
- Default Value: False

#### enable - Arguments

##### servicename
- Description: The name of the service to enable.
- Required Value: True
- Default Value: None

##### enable
- Description: Boolean flag to indicate if you want to enable a specific service
- Required Value: False
- Default Value: False

#### remove - Arguments

##### servicename
- Description: The name of the service to remove.
- Required Value: True
- Default Value: None

##### remove
- Description: Boolean flag to indicate if you want to remove a specific service
- Required Value: False
- Default Value: False

#### dumpstate - Arguments

##### dumpstate
- Description: Boolean flag to indicate if you want to dump the current state of launchd services
- Required Value: False
- Default Value: False

## Usage

```
xpc_service -print
xpc_service -print -servicename com.itsafeature.test
xpc_service -list
xpc_service -list -servicename com.itsafeature.test

```


## Detailed Summary

This command uses the `xpc_pipe_routine` function to send XPC messages to `launchd`.: