#import <Foundation/Foundation.h>
#include "dyld_inject_darwin.h"

// https://github.com/sfsam/Itsycal/blob/11e6e9d265379a610ef103850995e280873f9505/Itsycal/MoLoginItem.m
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wdeprecated-declarations"


bool openUsingLSWith(NSString *path, NSDictionary *env, bool hide) {
    FSRef appFSURL;
    OSStatus stat = FSPathMakeRef((const UInt8 *)[path UTF8String], &appFSURL, NULL);
    
    if (stat != errSecSuccess) {
        return false;
    }
    
    LSApplicationParameters appParam;
    appParam.version = 0;
    
    if (hide) {
        appParam.flags = kLSLaunchAndHide;
    } else {
        appParam.flags = kLSLaunchDefaults;
    }

    appParam.application = &appFSURL;
    
    appParam.argv = NULL;
    appParam.environment = (__bridge CFDictionaryRef)env;
    appParam.asyncLaunchRefCon = NULL;
    appParam.initialEvent = NULL;
    CFArrayRef array = (__bridge CFArrayRef)@[];
    stat = LSOpenURLsWithRole(array, kLSRolesAll, NULL, &appParam, NULL, 0);
    if (stat != errSecSuccess) {
        return false;
    }
    return true;
}

int dyld_inject(char *app, char *dylib, int hide) {
    @try {
        NSString *appPath = [NSString stringWithCString:app encoding:NSUTF8StringEncoding];
        NSString *dylibPath = [NSString stringWithCString:dylib encoding:NSUTF8StringEncoding];
        
        NSDictionary *env = @{@"DYLD_INSERT_LIBRARIES":dylibPath};
        
        bool shouldHide = false;
        if (hide == 1) {
            shouldHide = true;
        }
        
        bool success = openUsingLSWith(appPath, env, shouldHide);
        if (success != true) {
            return -1;
        }
        return 0;
    } @catch (NSException *exception) {
        return -1;
    }
}