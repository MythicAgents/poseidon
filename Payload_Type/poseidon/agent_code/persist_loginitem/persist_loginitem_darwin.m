#import <Foundation/Foundation.h>
#include "persist_loginitem_darwin.h"

// https://github.com/sfsam/Itsycal/blob/11e6e9d265379a610ef103850995e280873f9505/Itsycal/MoLoginItem.m
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wdeprecated-declarations"


int persist_loginitem(char *path, char *name, int global) {
    LSSharedFileListRef loginItemsRef = NULL;
    @try {
        NSString *pathString = [NSString stringWithUTF8String:path];
        NSString *nameString = [NSString stringWithUTF8String:name];
        if (global) {
            loginItemsRef = LSSharedFileListCreate(NULL, kLSSharedFileListGlobalLoginItems, NULL);
        } else {
            loginItemsRef = LSSharedFileListCreate(NULL, kLSSharedFileListSessionLoginItems, NULL);
        }
        
        if (loginItemsRef) {
            CFURLRef url = (__bridge CFURLRef)[NSURL fileURLWithPath:pathString];
            LSSharedFileListItemRef item = LSSharedFileListInsertItemURL(loginItemsRef, kLSSharedFileListItemLast, (__bridge CFStringRef)(nameString), NULL, url, NULL, NULL);
            if (item != NULL) {
                CFRelease(item);
                return 1;
            } else {
                return -1;
            }
        } else {
            return -1;
        }
        
    } @catch (NSException *exception) {
        NSLog(@"%@", exception);
        return -1;
    }
}