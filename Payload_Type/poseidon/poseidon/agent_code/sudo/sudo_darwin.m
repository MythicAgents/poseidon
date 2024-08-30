#import <Foundation/Foundation.h>
#import <Security/Security.h>
#import "sudo_darwin.h"


const char* sudo_poseidon(char* username, char* password, char* promptText, char* promptIcon, char* command, char** args, int* fd){
    AuthorizationRef authRef = 0;
    OSStatus status = 0;
    char* rightName = "allow";
    AuthorizationItem   environment[4] = {{NULL, 0, NULL, 0}, {NULL, 0, NULL, 0}, {NULL, 0, NULL, 0}, {NULL, 0, NULL, 0}};
    int numItems = 0;
    if ( strlen(username) > 0 ) {
        AuthorizationItem item = { kAuthorizationEnvironmentUsername, strlen(username), (char*)username, 0 };
        environment[numItems++] = item;
    }
    if ( strlen(password) > 0 ) {
        AuthorizationItem passItem = { kAuthorizationEnvironmentPassword, strlen(password), (char*)password, 0 };
        environment[numItems++] = passItem;
    }
    if ( strlen(promptText) > 0 ){
        AuthorizationItem promptItem = { kAuthorizationEnvironmentPrompt, strlen(promptText), (char*)promptText, 0 };
        environment[numItems++] = promptItem;
    }
    if ( strlen(promptIcon) > 0 ){
        AuthorizationItem iconItem = { kAuthorizationEnvironmentIcon, strlen(promptIcon), (char*)promptIcon, 0 };
        environment[numItems++] = iconItem;
    }
    AuthorizationItem right = {NULL, 0, NULL, 0};
    right.name = rightName;
    right.valueLength = 0;
    right.value = 0;
    AuthorizationRights rightSet = { 1, &right };
    AuthorizationRights environmentSet = { numItems, environment };
    status = AuthorizationCreate(NULL, &environmentSet, kAuthorizationFlagDefaults, &authRef);
    if (status != noErr) {
        return [[[NSString alloc] initWithFormat:@"Error: %d. Cannot create authorization reference.", status] UTF8String];
    }
    AuthorizationFlags flags = kAuthorizationFlagDefaults | kAuthorizationFlagExtendRights | kAuthorizationFlagPreAuthorize;// | kAuthorizationFlagInteractionAllowed; //<- Just for debugging, will display the OS auth dialog if needed!!!
    if( strlen(password) == 0 ){
        // only allow user authorization if we don't know the password
        flags |= kAuthorizationFlagInteractionAllowed;
    }
    status = AuthorizationCopyRights(authRef, &rightSet, &environmentSet, flags, NULL );
    if (status != errAuthorizationSuccess) {
        AuthorizationFree(authRef,kAuthorizationFlagDestroyRights);
        if(status == errAuthorizationCanceled){
            return "Error. User cancelled";
        }
        return [[[NSString alloc] initWithFormat:@"Error: %d. Cannot copy authorization reference.", status] UTF8String];
    }
    //char *args[] = {NULL};
    FILE *pipe = NULL;

    status = AuthorizationExecuteWithPrivileges(authRef, command, kAuthorizationFlagDefaults, args, &pipe);

    if (status == errAuthorizationSuccess) {
        AuthorizationFree(authRef,kAuthorizationFlagDestroyRights);
        *fd = fileno(pipe);
        return "";
        /*
        // Print to standard output
        char readBuffer[128];
        for (;;) {
            int bytesRead = read(fileno(pipe), readBuffer, sizeof(readBuffer));
            if (bytesRead < 1) break;
            write(fileno(stdout), readBuffer, bytesRead);
        }
        */
    } else {
        AuthorizationFree(authRef,kAuthorizationFlagDestroyRights);
        if( status == errAuthorizationToolExecuteFailure){
            return "Error: Specified program could not be executed. Make sure you supplied the full absolute path.";
        }
        return [[[NSString alloc] initWithFormat:@"Error: %d. Failed to execute with privs.", status] UTF8String];
    }
}