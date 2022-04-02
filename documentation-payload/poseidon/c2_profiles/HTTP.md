+++
title = "HTTP"
chapter = false
weight = 102
+++

## Summary


The `poseidon` agent uses HTTP POST messages for getting tasking and sending responses. The GET Query parameter is not used. This is done as an optimization step so that even when checking for tasking, we can pass along messages from linked agents or SOCKS data. 


### Profile Option Deviations

