#import <Foundation/Foundation.h>
#import <IOKit/pwr_mgt/IOPMLib.h>
#include "caffeinate_wrapper_darwin.h"

char* caffeinate(int enable) {
    @try {
            IOPMAssertionLevel newLevel = kIOPMAssertionLevelOn;
            if(enable == 0){
                newLevel = kIOPMAssertionLevelOff;
            }
            CFStringRef assertionName = CFStringCreateWithCString(NULL, "caffeinate", kCFStringEncodingUTF8);
            IOPMAssertionID assertionID;
            IOReturn status = IOPMAssertionCreateWithName(kIOPMAssertionTypePreventSystemSleep, newLevel, assertionName, &assertionID);
            if(status == kIOReturnSuccess){
                return "Successfully adjusted caffeinate status";
            } else {
                NSString* fmtString = [NSString stringWithFormat:@"Failed to set status: %d", status];
                return [fmtString UTF8String];
            }
    } @catch (NSException *exception) {
        return [[exception reason] UTF8String];
    }
    
}