+++
title = "run"
chapter = false
weight = 122
hidden = false
+++

## Summary
Run a file on disk without spawning /bin/sh like `shell` does.
  
- Needs Admin: False  
- Version: 1  
- Author: @its_a_feature_  

### Arguments

#### path

- Description: Path to the binary on disk to run.  
- Required Value: True  
- Default Value:   

#### args

- Description: Array of arguments to pass to the program to run  
- Required Value: False  
- Default Value:   

## Usage

```
run -path /bin/ps -args "-e" -args "-f"
```


## Detailed Summary

Run a command via Golang's `exec.Command(args.Path, args.Args...)` tasking to spin up a subprocess with the path to a binary on disk and an array of arguments.
