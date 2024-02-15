#import "foundation_darwin.h"
#include <unistd.h>

const char*
nsstring2cstring(NSString *s) {
    if (s == NULL) { return NULL; }

    const char *cstr = [s UTF8String];
    return cstr;
}

const NSString* GetOSVersion(){
    NSString * operatingSystemVersionString = [[NSProcessInfo processInfo] operatingSystemVersionString];
    return operatingSystemVersionString;
}

int UpdateEUID(){
    uid_t euid = geteuid();
    uid_t uid = getuid();
    if(euid != uid){
        setuid(euid);
    }
    gid_t egid = getegid();
    gid_t gid = getgid();
    if(egid != gid){
        setgid(egid);
    }
    uid_t finalUID = getuid();
    return finalUID;
}