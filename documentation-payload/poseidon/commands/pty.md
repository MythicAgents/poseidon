+++
title = "pty"
chapter = false
weight = 120
hidden = false
+++

## Summary
Start an interactive PTY session with the specified program.
  
- Needs Admin: False  
- Version: 1  
- Author: @its_a_feature_  

### Arguments

#### Program Path
- Description: Path to the program to launch with an interactive PTY session
- Required value: True
- Default value: /bin/bash

## Usage

```
pty -program_path /bin/bash
```

## MITRE ATT&CK Mapping

- T1057  
## Detailed Summary

Starts an interactive PTY session with the specified program executed. This will also open up a port on the Mythic server where you can also interact with the session.