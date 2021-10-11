+++
title = "persist_loginitem"
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

#### Path

- Description: Path to the binary to execute at login
- Required Value: True
- Default Value: None

#### Name

- Description: The name that is displayed in the Login Items section of the Users & Groups preferences pane
- Required Value: True
- Default Value: None

#### Global

- Description: Set this to true if the login item should be installed for all users. This requires administrative privileges
- Required Value: True
- Default Value: True

## Usage
```
persist_loginitem
```

## Detailed Summary

This function uses ObjectiveC API calls to set a new login item. Admin privileges are required if Global is set to True.