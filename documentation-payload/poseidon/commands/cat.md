+++
title = "cat"
chapter = false
weight = 100
hidden = false
+++

## Summary
Cat a file via golang functions. If the file size is greater than 5 * CHUNK_SIZE (typically 512kb) then cat won't return the contents of the file and will instruct you to download the file instead.
 
- Needs Admin: False  
- Version: 1  
- Author: @xorrior  

### Arguments

#### path

- Description: path to file (no quotes required)  
- Required Value: True  
- Default Value: None  

## Usage

```
cat /path/to/file
cat -path /path/to/file
```

## MITRE ATT&CK Mapping

- T1005  
## Detailed Summary

