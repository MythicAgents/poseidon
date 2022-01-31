+++
title = "persist_launchd"
chapter = false
weight = 120
hidden = false
+++

## Summary
Install launchd persistence

- Needs Admin: False  
- Version: 1  
- Author: @xorrior 

### Arguments

#### Args

- Description: List of arguments to execute in the ProgramArguments section of the PLIST
- Required Value: True
- Default Value: None

#### KeepAlive

- Description: When this value is set to true, Launchd will restart the daemon if it dies
- Required Value: False
- Default Value: True

#### RunAtLoad

- Description: When this value is set to true, Launchd will immediately start the daemon/agent once it has been registered
- Required Value: False
- Default Value: False

#### Label

- Description: The label for launchd persistence
- Required Value: True
- Default Value: com.apple.mdmupdateagent

#### LaunchPath

- Description: Path to save the new plist
- Required Value: True
- Default Value: None

#### LocalAgent

- Description: Should be a local user launch agent?
- Required Value: True
- Default Value: True

## Usage 
```
persist_launchd
```

## MITRE ATT&CK Mapping 

- T1159
- T1160

## Detailed Summary

Create a launch agent or daemon plist file and save it to ~/Library/LaunchAgents or /Library/LaunchDaemons. For additional information on launch agent parameters please visit: https://developer.apple.com/library/archive/documentation/MacOSX/Conceptual/BPSystemStartup/Chapters/CreatingLaunchdJobs.html