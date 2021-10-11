#import <Foundation/Foundation.h>
#include "spawn_libinject_darwin.h"

// https://github.com/sfsam/Itsycal/blob/11e6e9d265379a610ef103850995e280873f9505/Itsycal/MoLoginItem.m
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wdeprecated-declarations"


int spawn_libinject(char *app, char *dylib, char *args, int hide) {
    @try {
        NSString *appPath = [NSString stringWithCString:app encoding:NSUTF8StringEncoding];
        NSString *dylibPath = [NSString stringWithCString:dylib encoding:NSUTF8StringEncoding];
        
        NSDictionary *env = @{@"DYLD_INSERT_LIBRARIES":dylibPath};
        
        FSRef appFSURL;
        OSStatus stat = FSPathMakeRef((const UInt8 *)[appPath UTF8String], &appFSURL, NULL);
        
        if (stat != errSecSuccess) {
            return 1;
        }
        
        LSApplicationParameters appParam;
        appParam.version = 0;
        if (hide == 1) {
            appParam.flags = kLSLaunchAndHide;
        } else {
            appParam.flags = kLSLaunchDefaults;
        }
        appParam.application = &appFSURL;
        if (strlen(args) > 0) {
            NSString *nsArgString = [NSString stringWithCString:args encoding:NSUTF8StringEncoding];
            NSArray *cArgs = [nsArgString componentsSeparatedByString:@" "];
            CFMutableArrayRef array = CFArrayCreateMutable(kCFAllocatorDefault, (CFIndex)cArgs.count, NULL);
            
            for (int i=0; i < (CFIndex)cArgs.count; i++) {
                CFStringRef str = CFBridgingRetain(cArgs[i]);
                CFArrayAppendValue(array, str);
            }
            
            appParam.argv = array;
            
        } else {
            appParam.argv = NULL;
        }
        
        appParam.environment = (__bridge CFDictionaryRef)env;
        appParam.asyncLaunchRefCon = NULL;
        appParam.initialEvent = NULL;
        CFArrayRef arr = (__bridge CFArrayRef)@[];
        stat = LSOpenURLsWithRole(arr, kLSRolesAll, NULL, &appParam, NULL, 0);
        if (stat != errSecSuccess) {
            return 1;
        }
        return 0;
    } @catch (NSException *exception) {
        return 1;
    }
}