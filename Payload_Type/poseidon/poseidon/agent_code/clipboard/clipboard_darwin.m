#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>
#import <AppKit/NSWorkspace.h>
#import <AppKit/NSPasteboard.h>
#import "clipboard_darwin.h"

const char* getClipboard(int argc, char** argv){

    NSString* allTypes = [[NSString alloc] initWithUTF8String:"*"];
    NSMutableArray *inputs = [[NSMutableArray alloc] init];
    for(int i = 0; i < argc; i++){
        [inputs addObject:[[NSString alloc] initWithUTF8String:argv[i]]];
    }
    NSPasteboard *pb = [NSPasteboard generalPasteboard];
    NSArray *types = [pb types];
    NSMutableDictionary *clipboard = [[NSMutableDictionary alloc] init];
    for(int i = 0; i < [types count]; i++){
        if( [inputs containsObject:types[i] ] || [inputs containsObject:allTypes] ){
            [clipboard setValue:[[pb dataForType:types[i]] base64EncodedStringWithOptions:0 ] forKey:types[i]];
        } else {
            [clipboard setValue:@"" forKey:types[i]];
        }
    }
    NSData* clipboardJSONData = [NSJSONSerialization dataWithJSONObject:clipboard options:NSJSONWritingPrettyPrinted error:nil];
    NSString* clipboardJSONString = [[NSString alloc] initWithData:clipboardJSONData encoding:NSUTF8StringEncoding];
    return [clipboardJSONString UTF8String];
}