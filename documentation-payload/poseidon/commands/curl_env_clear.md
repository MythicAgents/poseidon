+++
title = "curl_env_clear"
chapter = false
weight = 103
hidden = false
+++

## Summary
Clear curl-specific environment variables

- Needs Admin: False  
- Version: 1  
- Author: @its_a_feature_  

### Arguments

#### clearEnv

- Description: Array of env names to remove.  
- Required Value: false  
- Default Value: None

#### clearAll

- Description: Boolean flag to clear all env variables for this callback.  
- Required Value: false  
- Default Value: None  

## Usage

```
curl_env_clear -clearEnv TOKEN -clearEnv URL
curl_env_clear -clearAll
```


## Detailed Summary

This command manages a state of internal environment variables for the `curl` command to leverage.