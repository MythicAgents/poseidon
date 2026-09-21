+++
title = "md5sum"
chapter = false
weight = 100
hidden = false
+++

## Summary

Calculate the MD5 hash of a file.

- Needs Admin: False  
- Version: 1  
- Author: @ice-wzl  

### Arguments

#### path

- Description: Path to the file to hash  
- Required Value: True  
- Default Value: None  

## Usage

```text
md5sum /path/to/file
```

## Detailed Summary

`md5sum` reads the specified file and returns its hexadecimal MD5 digest followed by the supplied file path. The command does not modify the file.
