+++
title = "execute_library"
chapter = false
weight = 103
hidden = false
+++

## Summary
Load a dylib from disk and execute a function from within it.

- Needs Admin: False  
- Version: 1  
- Author: @its_a_feature_  

### Arguments

#### function_name

- Description: Name of the function within the dylib to be executed
- Required Value: True  
- Default Value: None

#### file_path
- Description: Path of the file on disk to load (or path of where to upload file on disk)
- Required Value: True
- Default Value: None

#### args
- Description: Array of string args to pass to the function
- Required Value: True
- Default Value: None

#### file_id
- Description: File UUID of newly uploaded file (if uploading new file)
- Required Value: False
- Default Value: None

## Usage

```
execute_library -file_path /Users/itsafeature/Desktop/evil.dylib -function_name evil -args a -args "something else" -args "blah"
```


## Detailed Summary

This command uses dlopen and dylsim to open a library on disk and execute a function within it. 
This is used instead of NSCreateFileObjectImageFromMemory because that function will write out your in-memory library first to disk in a known location, which makes it more suspicious.