+++
title = "curl_env_set"
chapter = false
weight = 103
hidden = false
+++

## Summary
Sets curl-specific environment variables for use with the `curl` command.

- Needs Admin: False  
- Version: 1  
- Author: @its_a_feature_  

### Arguments

#### setEnv

- Description: Array of environment variables to set for curl command in the format of `Key=Value`.
- Required Value: True  
- Default Value: None

## Usage

```
curl_env_set -setEnv TOKEN=ejyaskdj -setEnv URL=https://mydomain.com
```


## Detailed Summary

This command sets/updates environment variables for use with `curl`.