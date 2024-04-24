//
//  launchdXPC.c
//  Created by Patrick Wardle
//  Ported from code by Jonathan Levin
//

#include <stdio.h>
#import <dlfcn.h>
#import "launchdXPC_darwin.h"
#import <Foundation/Foundation.h>
#import <ApplicationServices/ApplicationServices.h>

NSDictionary * infoForPID(pid_t pid)
{
    NSDictionary *ret = nil;
    ProcessSerialNumber psn = { 0, 0 };
    if (GetProcessForPID(pid, &psn) == noErr) {
        CFDictionaryRef cfDict = ProcessInformationCopyDictionary(&psn,kProcessDictionaryIncludeAllInformationMask);
        ret = [NSDictionary dictionaryWithDictionary:(__bridge NSDictionary *)cfDict];
        CFRelease(cfDict);
    }
    return ret;
}

int getSubmittedPid(int pid) {
    NSDictionary* info =  infoForPID(pid);
    //NSLog(@"info: %@", info);

    long long temp = [[info objectForKey:@"ParentPSN"] longLongValue];
    long long hi = (temp >> 32) & 0x00000000FFFFFFFFLL;
    long long lo = (temp >> 0) & 0x00000000FFFFFFFFLL;
    ProcessSerialNumber parentPSN = {(unsigned long)hi, (unsigned long)lo};

    NSDictionary* parentDict = (__bridge NSDictionary*)ProcessInformationCopyDictionary (&parentPSN, kProcessDictionaryIncludeAllInformationMask);
    //NSLog(@"real parent info: %@", parentDict);
    //NSLog(@"real parent pid: %@", parentDict[@"pid"]);

    pid_t p = 0;
    GetProcessPID(&parentPSN, &p);
    //NSLog(@"real parent pid: %d", p);

    return p;
}

#define ROUTINE_DUMP_PROCESS  0x2c4

//function definition
NSMutableDictionary* parse(NSString* data);

//hit up launchd (via XPC) to get process info
NSString* getSubmittedByPlist(unsigned long pid)
{
    //proc info
    NSDictionary* processInfo = nil;

    //xpc dictionary
    // passed to launchd to get proc info
    xpc_object_t procInfoRequest = NULL;

    //shared memory for XPC
    xpc_object_t sharedMemory = NULL;

    //xpc (out) dictionary
    // don't really contain response, but will have (any) errors
    xpc_object_t __autoreleasing response = NULL;

    //result
    int result = 0;

    //dylib handle
    void *handle = NULL;

    //function pointer to 'xpc_pipe_interface_routine'
    static int(*xpc_pipe_interface_routine_FP)(xpc_pipe_t, int, xpc_object_t, xpc_object_t*, int) = NULL;

    //(xpc) error
    int64_t xpcError = 0;

    //global data
    struct xpc_global_data* globalData = NULL;

    //init dictionary
    procInfoRequest = xpc_dictionary_create(NULL,NULL,0);

    //init buffer
    // size from reversing launchctl
    size_t processInfoLength = 0x100000;
    vm_address_t processInfoBuffer = 0;

    //bytes written via XPC call
    uint64_t bytesWritten = 0;

    //alloc buffer
    if(noErr != vm_allocate(mach_task_self(), &processInfoBuffer, processInfoLength, 0xf0000003))
    {
        //bail
        goto bail;
    }

    sharedMemory = xpc_shmem_create((void*)processInfoBuffer, processInfoLength);
    if(0 == sharedMemory)
    {
        //bail
        goto bail;
    }

    //add pid to request
    xpc_dictionary_set_int64(procInfoRequest, "pid", pid);

    //add shared memory object
    xpc_dictionary_set_value(procInfoRequest, "shmem", sharedMemory);

    //grab from global data structure
    // contains XPC bootstrap pipe (launchd)
    globalData = (struct xpc_global_data *)_os_alloc_once_table[1].ptr;

    //open XPC lib
    handle = dlopen("/usr/lib/system/libxpc.dylib", RTLD_LAZY);
    if(NULL == handle)
    {
        //bail
        goto bail;
    }

    xpc_pipe_interface_routine_FP = dlsym(handle, "_xpc_pipe_interface_routine");
    if(NULL == xpc_pipe_interface_routine_FP)
    {
        //bail
        goto bail;
    }

    //request process info
    result = xpc_pipe_interface_routine_FP((__bridge xpc_pipe_t)(globalData->xpc_bootstrap_pipe), ROUTINE_DUMP_PROCESS, procInfoRequest, &response, 0x0);
    if(0 != result)
    {
        //error
        goto bail;
    }

    //check for other error(s)
    xpcError = xpc_dictionary_get_int64(response, "error");
    if(0 != xpcError)
    {
        //error
        //printf("error: %llx\n", xpcError);
        goto bail;
    }

    //get number of bytes written (to shared memory)
    bytesWritten = xpc_dictionary_get_uint64(response, "bytes-written");
    //parse
    processInfo = parse([[NSString alloc] initWithBytes:(const void *)processInfoBuffer length:bytesWritten encoding:NSUTF8StringEncoding]);

bail:

    //free buffer
    if(0 != processInfoBuffer)
    {
        //free
        vm_deallocate(mach_task_self(), processInfoBuffer, processInfoLength);
        processInfoBuffer = 0;
    }
    if (processInfo){
        return processInfo[@"path"];
    }
    return @"";

    //return processInfo;
}


//parse proc info
NSMutableDictionary* parse(NSString* data)
{
    //pool
    @autoreleasepool {
        NSData* plistData = [data dataUsingEncoding:NSUTF8StringEncoding];
        NSString *error;
        NSPropertyListFormat format = NSPropertyListOpenStepFormat;
        NSDictionary* plist = [NSPropertyListSerialization propertyListWithData:plistData options:NSPropertyListImmutable format:&format error:&error];
        return plist;
    }
}
