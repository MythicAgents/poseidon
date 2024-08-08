#import <Foundation/Foundation.h>
#include "lsopen_darwin.h"

#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wdeprecated-declarations"

//referenced for base code - https://wojciechregula.blog/post/how-to-rob-a-firefox 
//expanded with dynamic argument handling, "_" var stomping, and an unset to DYLD_INSERT_LIBRARIES to prevent inherited injections into new processes

bool lsopen_call(NSString *path, bool hide, NSArray<NSString*> *arghhhs) {
    FSRef appFSURL;
    OSStatus stat = FSPathMakeRef((const UInt8 *)[path UTF8String], &appFSURL, NULL);

    NSDictionary *env = @{@"_":path};

    unsetenv("DYLD_INSERT_LIBRARIES");
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
    appParam.argv = (__bridge CFArrayRef) arghhhs;
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

int lsopen_init(char *app, int hide, char * argv[], int argc) {
    @try {
        NSString *appPath = [NSString stringWithCString:app encoding:NSUTF8StringEncoding];
        
        bool shouldHide = false;
        if (hide == 1) {
            shouldHide = true;
        }

        NSMutableArray *argarray = [NSMutableArray array];
        for (int i = 0; i < argc; i++) {
            NSString *str = [[NSString alloc] initWithCString:argv[i] encoding:NSUTF8StringEncoding];
            [argarray addObject:str];
        }

        NSRange rng = NSMakeRange(1, argc -1);
        NSArray* applicationargs = [argarray subarrayWithRange:rng];
        
        bool success = lsopen_call(appPath, shouldHide, applicationargs);
        if (success != true) {
            return -1;
        }
        return 0;
    } @catch (NSException *exception) {
        return -1;
    }
}