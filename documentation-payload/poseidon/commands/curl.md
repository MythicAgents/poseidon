+++
title = "curl"
chapter = false
weight = 103
hidden = false
+++

## Summary
Execute a single web request.

- Needs Admin: False  
- Version: 1  
- Author: @xorrior  

### Arguments

#### url

- Description: URL to request.  
- Required Value: True  
- Default Value: https://www.google.com  

#### method

- Description: Type of request  
- Required Value: True  
- Default Value: None  

#### headers

- Description: Array of headers in "Key: Value" entries. 
- Required Value: False  
- Default Value: None

#### body

- Description: String body contents to send.
- Required Value: False
- Default Value: None

## Usage

```
curl -url "https://www.google.com" -method "GET" -headers "User-Agent: Test" -headers "Content-Type: text/html"
curl -url $TARGET_URL/api/evil -method "GET" -headers "Authorization: Bearer $TOKEN" -headers "Content-Type: text/html"
```


## Detailed Summary

This command uses the Golang http.Client to perform a GET or POST request with optional arguments for request headers and body. 
If you set any curl-specific environment variables with `curl_env_set` then you can replace them inline like the second usage example.