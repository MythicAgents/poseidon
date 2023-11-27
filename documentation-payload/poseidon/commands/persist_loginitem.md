+++
title = "persist_loginitem"
chapter = false
weight = 120
hidden = false
+++

## Summary
Install loginitem persistence

- Needs Admin: False  
- Version: 1  
- Author: @xorrior 

### Arguments

#### Path

- Description: Path to the binary to execute at login
- Required Value: False
- Default Value: None

#### Name

- Description: The name that is displayed in the Login Items section of the Users & Groups preferences pane
- Required Value: False
- Default Value: None

#### Global

- Description: Set this to true if the login item should be installed for all users. This requires administrative privileges
- Required Value: False
- Default Value: True

#### list
- Description: Set this to true to list out all session and global login items
- Required Value: False
- Default Value: False

#### remove
- Description: Remove the specified loginitem persistence
- Required Value: False
- Default Value: False

## Usage
```
persist_loginitem
```

## Detailed Summary

This function uses ObjectiveC API calls to set a new login item. Admin privileges are required if Global is set to True.