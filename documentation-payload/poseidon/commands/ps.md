+++
title = "ps"
chapter = false
weight = 120
hidden = false
+++

## Summary
Get a process listing.
  
- Needs Admin: False  
- Version: 1  
- Author: @xorrior  

### Arguments

#### Regex Filter
- Description: Filter which processes get returned based on a regex filter
- Required value: False
- Default value: False

## Usage

```
ps
ps -regex_filter poseidon.*
```

## MITRE ATT&CK Mapping

- T1057  
## Detailed Summary

Obtain a list of running processes