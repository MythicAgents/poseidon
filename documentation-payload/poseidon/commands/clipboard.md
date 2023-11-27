+++
title = "clipboard"
chapter = false
weight = 103
hidden = false
+++

## Summary
Get a user's clipboard on macOS and send that data back based on which entries the user selects to capture.
This will always return the entire set of possible keys, but only the data for the selected entries.

- Needs Admin: False  
- Version: 1  
- Author: @its_a_feature_  

### Arguments

#### Read Types

- Description: An array of the names of entries to read the data for. A value of "*" means to read the data for all keys.
- Required Value: True  
- Default Value: ["public.utf8-plain-text"]

## Usage

```
clipboard -read "public.utf8-plain-text"
```


## Detailed Summary

This command uses the macOS NSPasteboard APIs to read the clipboard and return the base64 representation of the contents.
