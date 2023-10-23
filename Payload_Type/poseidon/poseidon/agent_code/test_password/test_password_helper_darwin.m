#include "test_password_helper_darwin.h"
#import <Foundation/Foundation.h>
#import <OpenDirectory/OpenDirectory.h>

char* testPassword(char* user, char* password){
    NSMutableString* output = [[NSMutableString alloc] initWithString:@""];
    ODSession* session = [ODSession defaultSession];
    ODNode *node = [ODNode nodeWithSession:session type:kODNodeTypeAuthentication error:nil];
    NSError *err = NULL;
    ODRecord *userRecord = [node recordWithRecordType:kODRecordTypeUsers name:[[NSString alloc] initWithFormat:@"%s", "test"] attributes:nil error:&err];
    if(err != NULL){
        return [err.localizedDescription UTF8String];
    }
    bool success = [userRecord verifyPassword:[[NSString alloc] initWithFormat:@"%s", "password"] error:&err];
    if(err != NULL){
        return [err.localizedDescription UTF8String];
    }
    if(success){
        [output appendString:@"Authentication: Success!\n"];
    } else {
        [output appendString:@"Authentication: Failure!\n"];
    }
    return [output UTF8String];
}