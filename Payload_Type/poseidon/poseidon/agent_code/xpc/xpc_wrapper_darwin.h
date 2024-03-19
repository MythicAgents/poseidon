#ifndef _XPC_WRAPPER_H_
#define _XPC_WRAPPER_H_

#include <stdlib.h>
#include <stdio.h>
#include <xpc/xpc.h>
#include <xpc/connection.h>
#include <sys/utsname.h>
#include <launch.h>
#include <errno.h>

extern xpc_type_t TYPE_ERROR;
extern xpc_type_t TYPE_ARRAY;
extern xpc_type_t TYPE_DATA;
extern xpc_type_t TYPE_DICT;
extern xpc_type_t TYPE_INT64;
extern xpc_type_t TYPE_UINT64;
extern xpc_type_t TYPE_STRING;
extern xpc_type_t TYPE_UUID;
extern xpc_type_t TYPE_BOOL;
extern xpc_type_t TYPE_DATE;
extern xpc_type_t TYPE_FD;
extern xpc_type_t TYPE_CONNECTION;
extern xpc_type_t TYPE_NULL;
extern xpc_type_t TYPE_SHMEM;
extern xpc_object_t ERROR_CONNECTION_INVALID;
extern xpc_object_t ERROR_CONNECTION_INTERRUPTED;
extern xpc_object_t ERROR_CONNECTION_TERMINATED;

extern xpc_connection_t XpcConnect(char *, uintptr_t, int);
extern void XpcSendMessage(xpc_connection_t, xpc_object_t, bool, bool);
extern void XpcArrayApply(uintptr_t, xpc_object_t);
extern void XpcDictApply(uintptr_t, xpc_object_t);
extern void XpcUUIDGetBytes(void *, xpc_object_t);
extern xpc_object_t XpcLaunchdListServices(char *);
extern char* XpcLaunchdPrint(char *);
extern xpc_object_t XpcLaunchdServiceControl(char *, int);
extern xpc_object_t XpcLaunchdServiceControlEnableDisable(char *, int);
extern xpc_object_t XpcLaunchdSubmitJob(char *, char *, int);
extern xpc_object_t XpcLaunchdRemove(char *);
extern xpc_object_t XpcLaunchdAsUser(char *program, int uid);
extern xpc_object_t XpcLaunchdGetManagerUID(void);
extern char* XpcLaunchdDumpState(void);
extern xpc_object_t XpcLaunchdLoadPlist(char *, int);
extern char* XpcLaunchdGetProcInfo(unsigned long);
extern xpc_object_t XpcLaunchdUnloadPlist(char *);

extern void *objc_retain (void *);
extern int xpc_pipe_routine (xpc_object_t *, xpc_object_t *, xpc_object_t **);
extern char *xpc_strerror (int);
extern int csr_check (int what);

// This is undocumented, but sooooo useful :)
extern mach_port_t xpc_dictionary_copy_mach_send(xpc_object_t, const char *key);


// Some of the routine #s launchd recognizes. There are quite a few subsystems
// (stay tuned for MOXiI 2 - list is too long for now)

#define ROUTINE_DEBUG		0x2c1	// 705
#define ROUTINE_SUBMIT		100
/*
= "<dictionary: 0x600001b00960> { count = 5, transaction: 0, voucher = 0x0, contents =
	"subsystem" => <uint64: 0x25b6b6d3218c52d3>: 7
	"routine" => <uint64: 0x25b6b6d3218a62d3>: 100
	"handle" => <uint64: 0x25b6b6d3218c22d3>: 0
	"request" => <dictionary: 0x60000010d080> { count = 1, transaction: 0, voucher = 0x0, contents =
		"SubmitJob" => <dictionary: 0x60000010d7a0> { count = 4, transaction: 0, voucher = 0x0, contents =
			"KeepAlive" => <bool: 0x7ff846203120>: true
			"Label" => <string: 0x600002b63b10> { length = 24, contents = "com.itsafeature.testuser" }
			"ProgramArguments" => <array: 0x600002b60e40> { count = 0, capacity = 0, contents =
			}
			"Program" => <string: 0x600002b63de0> { length = 39, contents = "/Users/itsafeature/Desktop/poseidon.bin" }
		}
	}
	"type" => <uint64: 0x25b6b6d3218c52d3>: 7
}"
*/
#define ROUTINE_BLAME		0x2c3 	// 707
#define ROUTINE_DUMP_PROCESS	0x2c4	// 708
#define ROUTINE_RUNSTATS	0x2c5	// 709
#define ROUTINE_LOAD		0x320	// 800
#define ROUTINE_UNLOAD		0x321	// 801
#define ROUTINE_LOOKUP		0x324
#define ROUTINE_ENABLE		0x328	// 808
#define ROUTINE_DISABLE		0x329   // 809
#define ROUTINE_STATUS		0x32b   // 811

#define ROUTINE_KILL		0x32c
#define ROUTINE_VERSION		0x33c
#define ROUTINE_PRINT_CACHE	0x33c
#define ROUTINE_PRINT		0x33c	// also VERSION.., cache..
/* launchctl print-cache
 = "<dictionary: 0x6000036a92c0> { count = 6, transaction: 0, voucher = 0x0, contents =
	"subsystem" => <uint64: 0x20b97be030b61999>: 3
	"handle" => <uint64: 0x20b97be030b62999>: 0
	"shmem" => <shmem: 0x600000775bf0>: 20971520 bytes (5120 pages)
	"routine" => <uint64: 0x20b97be03085e999>: 828
	"type" => <uint64: 0x20b97be030b63999>: 1
	"cache" => <bool: 0x7ff84432b120>: true
}"
// launchctl print gui/501/com.itsafeature.testing
xpc_dictionary_get_string ( dictionary@0x600001b00960,"session")
 = "<dictionary: 0x600001b00960> { count = 6, transaction: 0, voucher = 0x0, contents =
	"subsystem" => <uint64: 0x25b6b6d3218c02d3>: 2
	"handle" => <uint64: 0x25b6b6d3219372d3>: 501
	"shmem" => <shmem: 0x600002b59b30>: 1048576 bytes (256 pages)
	"routine" => <uint64: 0x25b6b6d321a062d3>: 708
	"type" => <uint64: 0x25b6b6d3218ca2d3>: 8
	"name" => <string: 0x600002b58fc0> { length = 23, contents = "com.itsafeature.testing" }
}"
*/
#define ROUTINE_REBOOT_USERSPACE	803 // 10.11/9.0 only
#define ROUTINE_START		0x32d	// 813
#define ROUTINE_STOP		0x32e	// 814
#define ROUTINE_LIST		0x32f	// 815
#define ROUTINE_REMOVE 816
/*
 = "<dictionary: 0x600001b0d590> { count = 7, transaction: 0, voucher = 0x0, contents =
	"subsystem" => <uint64: 0x25b6b6d3218c12d3>: 3
	"handle" => <uint64: 0x25b6b6d3218c22d3>: 0
	"routine" => <uint64: 0x25b6b6d321bf22d3>: 816
	"type" => <uint64: 0x25b6b6d3218c52d3>: 7
	"name" => <string: 0x600002b5b1e0> { length = 23, contents = "com.itsafeature.testing" }
	"legacy" => <bool: 0x7ff846203120>: true
	"domain-port" => <mach send right: 0x600002443b80> { name = 74243, right = send, urefs = 208 }
}"
*/
#define ROUTINE_SETENV		0x333	// 819
#define ROUTINE_GETENV		0x334  // 820
#define ROUTINE_RESOLVE_PORT		0x336
#define ROUTINE_EXAMINE		0x33a
#define ROUTINE_LIMIT		0x339	// 825
#define ROUTINE_DUMP_DOMAIN	0x33c	// 828
#define ROUTINE_ASUSER	0x344	// 836, but really 835 maybe?
/*
 = "<dictionary: 0x600001b08960> { count = 5, transaction: 0, voucher = 0x0, contents =
	"subsystem" => <uint64: 0x25b6b6d3218c12d3>: 3
	"routine" => <uint64: 0x25b6b6d321b812d3>: 835
	"handle" => <uint64: 0x25b6b6d3218c22d3>: 0
	"uid" => <uint64: 0x25b6b6d3219372d3>: 501
	"type" => <uint64: 0x25b6b6d3218c32d3>: 1
}"
*/
#define ROUTINE_DUMP_STATE	0x342	// 834
/* launchctl dumpstate
xpc_dictionary_get_string ( dictionary@0x6000036ac0f0,"session")
 = "<dictionary: 0x6000036ac0f0> { count = 5, transaction: 0, voucher = 0x0, contents =
	"subsystem" => <uint64: 0x20b97be030b61999>: 3
	"handle" => <uint64: 0x20b97be030b62999>: 0
	"shmem" => <shmem: 0x6000006f10e0>: 20971520 bytes (5120 pages)
	"routine" => <uint64: 0x20b97be030820999>: 834
	"type" => <uint64: 0x20b97be030b63999>: 1
}"
*/
#define ROUTINE_DUMPJPCATEGORY		0x345	// was 346 in iOS 9
// the input type for xpc_uuid_create should be uuid_t but CGO instists on unsigned char *
// typedef uuid_t * ptr_to_uuid_t;
typedef unsigned char * ptr_to_uuid_t;
extern const ptr_to_uuid_t ptr_to_uuid(void *p);

// https://github.com/ProcursusTeam/launchctl/blob/main/xpc_private.h#L93C1-L104C3
int64_t xpc_user_sessions_enabled(void) __API_AVAILABLE(ios(16.0));
uint64_t xpc_user_sessions_get_foreground_uid(uint64_t) __API_AVAILABLE(ios(16.0));
enum {
	ENODOMAIN = 112,
	ENOSERVICE = 113,
	E2BIMPL = 116,
	EUSAGE = 117,
	EBADRESP = 118,
	EDEPRECATED = 126,
	EMANY = 133,
	EBADNAME = 140,
	ENOTDEVELOPMENT = 142,
	EWTF = 153,
};

#endif