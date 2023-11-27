+++
title = "clipboard_monitor"
chapter = false
weight = 103
hidden = false
+++

## Summary
Monitor a user's clipboard on macOS and send that data back every second if there's been a change.
This reports data back to the responses for that task as well as the keylog data.

- Needs Admin: False  
- Version: 1  
- Author: @its_a_feature_  

### Arguments

#### duration

- Description: The number of seconds to monitor the clipboard (a negative value means forever).  
- Required Value: True  
- Default Value: 

## Usage

```
clipboard_monitor -1
```


## Detailed Summary

This command uses the macOS NSPasteboard APIs to check if the pasteboard count has changed, and if so, returns the string representation of the contents. 
