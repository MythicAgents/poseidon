#import <Foundation/Foundation.h>
#import <Security/Security.h>
extern const char* sudo_poseidon(char* username, char* password, char* promptText, char* promptIcon, char* command, char**args, int* fd);