+++
title = "OPSEC"
chapter = false
weight = 10
pre = "<b>1. </b>"
+++

### Post-Exploitation Jobs
All poseidon commands execute in the context of a go routine or thread. These routines cannot be stopped once started. A "Stop" flag can be sent to them, but if the command itself isn't pausing periodically to check the status of the flag, then it can't be force-exited.

### Remote Process Injection
The libinject command will use the `process_set_tasks` function to get a task port (handle) of the target process. The shellcode payload used will load a specified dylib from disk into the target process. 

### Agent Compilation
There is currentlyno agent obfuscation.
