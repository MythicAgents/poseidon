#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>
#import <AppKit/NSWorkspace.h>
#import <AppKit/NSPasteboard.h>


extern char* monitorClipboard(long currentChangeCount);
extern long getClipboardCount();
extern char* getFrontmostApp();
extern void waitForTime();