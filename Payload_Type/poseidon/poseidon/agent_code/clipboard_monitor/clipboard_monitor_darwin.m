#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>
#import <AppKit/NSWorkspace.h>
#import <AppKit/NSPasteboard.h>
#import "clipboard_monitor_darwin.h"


@interface ActivateNotifications : NSObject
@property char* frontmost;
-(id) init;
-(void)updateFrontmost:(NSNotification *)notification;
-(NSString*)getFrontmostName;
@end

@implementation ActivateNotifications
-(id) init {
    if ((self = [super init])) {
        self.frontmost = [[[NSWorkspace sharedWorkspace].frontmostApplication localizedName] UTF8String];
        [[NSWorkspace sharedWorkspace].notificationCenter 	addObserver:self
                                                            selector:@selector(updateFrontmost:)
                                                            name:NSWorkspaceDidActivateApplicationNotification
                                                            object:nil];
    }

    return self;
}
-(void)updateFrontmost:(NSNotification *)notification
{
	self.frontmost = [[[[notification userInfo] objectForKey:NSWorkspaceApplicationKey] localizedName] UTF8String];
	//NSLog(@"Self.frontmost updated to: %s", self.frontmost);
	//NSLog(@"Reacting to notification %@ from object %@ with userInfo %@", notification, notification.object, notification.userInfo);
}
-(NSString*) getFrontmostName {
	//NSLog(@"Self.frontmost: %s", self.frontmost);
    return self.frontmost;
}
@end
ActivateNotifications* myNotifications = NULL;
char* monitorClipboard(long currentChangeCount){
	if(!myNotifications){
		//NSLog(@"myNotifications isn't set, setting it in monitorClipboard");
		myNotifications = [[ActivateNotifications alloc] init];
	}
    NSPasteboard *pb = [NSPasteboard generalPasteboard];
    long newChangeCount = pb.changeCount;
	if(newChangeCount != currentChangeCount){
		NSString* contents = [pb stringForType:NSPasteboardTypeString];
		return [contents UTF8String];
	} else {
		return "";
	}
}
long getClipboardCount(){
	NSPasteboard *pb = [NSPasteboard generalPasteboard];
    return pb.changeCount;
}
bool addedNewObserver = false;
char* getFrontmostApp(){
	if(!myNotifications){
		//NSLog(@"myNotifications isn't set, setting it in getFrontmostApp");
		myNotifications = [[ActivateNotifications alloc] init];
	}
	if( [myNotifications getFrontmostName] != NULL){
		return [myNotifications getFrontmostName];
	} else {
		return "";
	}
}
void waitForTime(){
    NSLog(@"called waitForTime");
	NSDate *now = [NSDate date];
	NSCalendar *currentCalendar = [NSCalendar currentCalendar];
	NSDate *nowPlusOneSecond = [currentCalendar dateByAddingUnit:NSCalendarUnitSecond
															value:5
															toDate:now
															options:NSCalendarMatchNextTime];
	[[NSRunLoop mainRunLoop] runUntilDate:nowPlusOneSecond];
}