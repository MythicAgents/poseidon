+++
title = "ledit"
chapter = false
weight = 100
hidden = false
+++

## Summary

Remove one or more lines from a text file by their one-based line numbers.

- Needs Admin: False  
- Version: 1  
- Author: @ice-wzl  

### Arguments

#### numbers

- Description: Comma-separated list of one-based line numbers to remove  
- Required Value: True  
- Default Value: None  

#### path

- Description: Path to the text file to modify  
- Required Value: True  
- Default Value: None  

## Usage

Remove lines 1, 5, and 10 from a file:

```text
ledit -numbers 1,5,10 -path /var/log/syslog
```

Use `cat` to display the file with line numbers before selecting lines to remove.

## MITRE ATT&CK Mapping

- T1685  

## Detailed Summary

`ledit` reads the specified file, removes each line listed in `numbers`, and writes the remaining content back to the same file. Line numbers are one-based. Both `path` and `numbers` are required, and every supplied line number must exist in the file.
