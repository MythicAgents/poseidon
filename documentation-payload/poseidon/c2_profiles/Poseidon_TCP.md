+++
title = "poseidon_tcp"
chapter = false
weight = 5
+++

## Overview
This C2 profile is a peer-to-peer (P2P) profile. This means that an agent with this profile does _not_ reach out to Mythic; instead, this profile opens a port on the target system and waits for incoming connections. You need to then have another agent use the `link_tcp` command to connect to this agent before issuing tasking. 

The Docker container for this profile does not "start" or "stop" since it's not going to be receiving any connections from agents. All connections are between two specific agents. This is why this profile doesn't have an option to start/stop or even configure it within the Mythic UI.
 
### C2 Workflow
{{<mermaid>}}
sequenceDiagram
    participant M as Mythic
    participant H as HTTP Container
    participant A as Agent1
    participant B as Agent2
    Note over B: Bind to port
    A ->>+ B: TCP Connect
    B -->>- A: Established Connection
    B ->>+ A: Checkin Message
    A ->>+ H: Forward Message
    H ->>+ M: Forward Message to Mythic
    M -->>- H: reply with new callback
    H -->>- A: reply with new callback
    A -->>- B: reply with new callback
    loop Egress Agent Get Tasking
    A ->>+ H: Get Tasking
    H ->>+ M: Forward Message
    M -->>- H: Tasking for A and B
    H -->>- A: Tasking for A and B
    A -->> B: Tasking for B
    B ->>+ A: Tasking Response for B
    A ->>+ H: Forward Tasking Response (A and B)
    H ->>+ M: Forward Tasking Response
    M -->>- H: Responses (A and B)
    H -->>- A: Responses (A and B)
    A -->>- B: Responses B
    end
    
{{< /mermaid >}}


## Configuration Options
There's no internal `poseidon_tcp` Docker container configuration that's required. The profile is used for inter-agent communication, so there's no agent messages that go through the `poseidon_tcp` Docker container.

### Profile Options
#### crypto type
Indicate if you want to use no crypto (i.e. plaintext) or if you want to use Mythic's aes256_hmac. Using no crypto is really helpful for agent development so that it's easier to see messages and get started faster, but for actual operations you should leave the default to aes256_hmac.

#### Port to Open
Number to specify the port number to callback to. This is split out since you don't _have_ to connect to the normal port (i.e. you could connect to http on port 8080). 

#### Kill Date
Date for the agent to automatically exit, typically the after an assessment is finished.

#### Perform Key Exchange
T or F for if you want to perform a key exchange with the Mythic Server. When this is true, the agent uses the key specified by the base64 32Byte key to send an initial message to the Mythic server with a newly generated RSA public key. If this is set to `F`, then the agent tries to just use the base64 of the key as a static AES key for encryption. If that key is also blanked out, then the requests will all be in plaintext.

## OPSEC

This profile opens a port on the host where the agent is running. 

## Development

