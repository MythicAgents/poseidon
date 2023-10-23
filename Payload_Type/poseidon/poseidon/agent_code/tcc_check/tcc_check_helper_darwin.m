#include "tcc_check_helper_darwin.h"
#import <Foundation/Foundation.h>
#import <ApplicationServices/ApplicationServices.h>

char* checkTCC(char* user){
    bool fullDiskAccess = false;
    bool desktopAccess = false;
    bool documentsAccess = false;
    bool downloadsAccess = false;
    NSString* userTCCPath;
    NSString* username;
    NSString* fdaQueryString = @"kMDItemDisplayName = *TCC.db";
    NSMutableString* output = [[NSMutableString alloc] initWithString:@""];
    NSString* suppliedUser = [[NSString alloc] initWithUTF8String:user];
    // https://github.com/MythicAgents/hermes/blob/main/Payload_Type/hermes/agent_code/Hermes/commands/get_execution_context.swift
    // https://cedowens.medium.com/spotlighting-your-tcc-access-permissions-ec6628d7a876
    // information about execution context environment variables from Justin Bui's & Cedric Owen's information
    [output appendString:@"Execution Context:\n"];
    NSDictionary* env = [NSProcessInfo processInfo].environment;
    NSString* bundleID = env[@"__CFBundleIdentifier"];
    [output appendFormat:@"__CFBundleIdentifier: %s\n", [bundleID UTF8String]];
    NSString* xpcService = env[@"XPC_SERVICE_NAME"];
    [output appendFormat:@"XPC_SERVICE_NAME: %s\n", [xpcService UTF8String]];
    NSString* packagePath = env[@"PACKAGE_PATH"];
    [output appendFormat:@"PACKAGE_PATH: %s\n", [packagePath UTF8String]];
    [output appendString:@"\n\nTCC Accesses:\n"];
    if(![suppliedUser isEqualToString:@""] ){
        username = [[NSString alloc] initWithFormat:@"%s", user];
    }else{
        username = NSUserName();
    }
    if( [username isEqualToString:@"root"] ){
        return "Currently the root user - must supply a username to check";
    } else {
        userTCCPath = [[NSString alloc] initWithFormat:@"/Users/%s/Library/Application Support/com.apple.TCC/TCC.db", [username UTF8String]];
    }
    // check for full disk access
    // https://github.com/MythicAgents/hermes/blob/main/Payload_Type/hermes/agent_code/Hermes/commands/fda_check.swift
    // pulled from Justin Bui's Hermes agent code
    MDQueryRef query = MDQueryCreate(kCFAllocatorDefault, (__bridge CFStringRef)fdaQueryString, nil, nil);
    if(query == NULL){
        [output appendString:@"Full Disk Access: unknown - failed to query\n"];
    } else {
        MDQueryExecute(query, kMDQuerySynchronous);
        for(int i = 0; i < MDQueryGetResultCount(query); i++){
            MDItemRef item = MDQueryGetResultAtIndex(query, i);
            NSString* path = CFBridgingRelease(MDItemCopyAttribute(item, kMDItemPath));
            if( [path hasSuffix:userTCCPath] ){
                fullDiskAccess = true;
            }
        }
        if(fullDiskAccess){
            [output appendString:@"Full Disk Access: true\n"];
        } else {
            [output appendString:@"Full Disk Access: false\n"];
        }
    }
    // https://github.com/MythicAgents/hermes/blob/main/Payload_Type/hermes/agent_code/Hermes/commands/tcc_folder_check.swift
    // pulled from Justin Bui's Hermes agent code
    NSString* queryFolderString = [[NSString alloc] initWithFormat:@"kMDItemKind = Folder -onlyin /Users/%s", [username UTF8String] ];
    query = MDQueryCreate(kCFAllocatorDefault, (__bridge CFStringRef)queryFolderString, nil, nil);
    if(query == NULL){
        [output appendString:@"Desktop Access: unknown - failed to query\n"];
        [output appendString:@"Documents Access: unknown - failed to query\n"];
        [output appendString:@"Downloads Access: unknown - failed to query\n"];
    } else {
        MDQueryExecute(query, kMDQuerySynchronous);
        for(int i = 0; i < MDQueryGetResultCount(query); i++){
            MDItemRef item = (MDItemRef) MDQueryGetResultAtIndex(query, i);
            NSString* path = CFBridgingRelease(MDItemCopyAttribute(item, kMDItemPath));
            if( [path isEqualToString:[[NSString alloc] initWithFormat:@"/Users/%s/Desktop", [username UTF8String]]] ){
                desktopAccess = true;
            }
            if( [path isEqualToString:[[NSString alloc] initWithFormat:@"/Users/%s/Documents", [username UTF8String]]] ){
                documentsAccess = true;
            }
            if( [path isEqualToString:[[NSString alloc] initWithFormat:@"/Users/%s/Downloads", [username UTF8String]]] ){
                downloadsAccess = true;
            }
        }
        if(desktopAccess){
            [output appendString:@"Desktop Access: true\n"];
        } else {
            [output appendString:@"Desktop Access: false\n"];
        }
        if(documentsAccess){
            [output appendString:@"Documents Access: true\n"];
        } else {
            [output appendString:@"Documents Access: false\n"];
        }
        if(downloadsAccess){
            [output appendString:@"Downloads Access: true\n"];
        } else {
            [output appendString:@"Downloads Access: false\n"];
        }
    }
    if( AXIsProcessTrusted() ){
        [output appendString:@"Accessibility Enabled: true\n"];
    } else {
        [output appendString:@"Accessibility Enabled: false\n"];
    }
    return [output UTF8String];
}