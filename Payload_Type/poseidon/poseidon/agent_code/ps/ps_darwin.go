//go:build darwin

package ps

/*
#cgo LDFLAGS: -framework AppKit -framework Foundation -framework ApplicationServices
#cgo CFLAGS: -x objective-c
#include "rdprocess_darwin.h"
#include <libproc.h>
#include <xpc/xpc.h>
#include <mach/task.h>
#include <objc/objc.h>
#import <dlfcn.h> // needed for dylsm
#import <Foundation/Foundation.h>
#import "launchdXPC_darwin.h"
const int INFO_SIZE = 136;

int getPPID(int pidOfInterest) {
    // Call proc_pidinfo and return nil on error
    struct proc_bsdinfo pidInfo;
    int InfoSize = proc_pidinfo(pidOfInterest, PROC_PIDTBSDINFO, 0, &pidInfo, INFO_SIZE);
    if(!InfoSize){
        return 1;
    }
    return pidInfo.pbi_ppid;
}
typedef int (*rpidFunc)(int);
int getResponsiblePid(int pidOfInterest) {
    // Get responsible pid using private Apple API
    const char* path = "/usr/lib/system/libquarantine.dylib";
    void* libquarHandle = dlopen(path, 2);
    void* rpidSym = dlsym(libquarHandle, "responsibility_get_pid_responsible_for_pid");
    if(rpidSym != NULL){
        int responsiblePid = ((rpidFunc)rpidSym)(pidOfInterest);
        if( responsiblePid == -1){
            //printf("Error getting responsible pid for process \(pidOfInterest). Setting to responsible pid to itself\n");
            return pidOfInterest;
        } else {
            return responsiblePid;
        }
    } else {
        return pidOfInterest;

    }
}
char* GetProcInfo(int pid) {
	RDProcess *p = [[RDProcess alloc] initWithPID:pid];
	NSString* source = @"unknown";
	int ppid = getPPID(pid);
	int responsiblePid = getResponsiblePid(pid);
	int submittedPid = getSubmittedPid(pid);
	int trueParent;
	if (submittedPid > 1) {
		trueParent = submittedPid;
		source = @"application_services";
	} else if (responsiblePid != pid) {
		trueParent = responsiblePid;
		source = @"responsible_pid";
	} else {
		trueParent = ppid;
		source = @"parent_process";
	}
	// Collect a plist if it caused this program to run
	NSString* plistNode = getSubmittedByPlist(pid);
	if ( [plistNode hasSuffix:@".plist"] ){
		source = @"launchd_xpc";
	}
	if(pid == 1){
		trueParent = 0; // make sure we don't set parent of pid 1 to pid 1 by accident
	}
    NSMutableDictionary *proc_details = [@{
        @"bundleID":p.bundleID ? : @"",
        @"args":p.launchArguments ? : @"",
        @"path":p.executablePath ? : @"",
        @"user":p.ownerUserName ? : @"",
        @"full_username":p.ownerFullUserName ? : @"",
        @"env": p.environmentVariables ? : @"",
        @"sandboxcontainer":p.sandboxContainerPath ? : @"",
		@"pid":[NSNumber numberWithInt:p.pid],
		@"scripting_properties":p.scriptingProperties ? : @"",
		@"name":p.processName ? : @"",
		@"ppid":[[NSNumber alloc] initWithInt:trueParent],
		@"source": source,
		@"responsible_pid": [[NSNumber alloc] initWithInt:responsiblePid],
		@"submitted_pid": [[NSNumber alloc] initWithInt:submittedPid],
		@"parent_pid": [[NSNumber alloc] initWithInt:ppid],
		@"backing_plist": plistNode ? : @"",
	}mutableCopy];

	NSError *error = nil;
    if ([NSJSONSerialization isValidJSONObject:proc_details]) {
        NSData* jsonData = [NSJSONSerialization dataWithJSONObject:proc_details options:NSJSONWritingPrettyPrinted error:&error];

        if (jsonData != nil && error == nil)
        {
        	NSString *jsonString = [[NSString alloc] initWithData:jsonData encoding:NSUTF8StringEncoding];

        	return [jsonString UTF8String];
        }
	}

	return "";
}
*/
import "C"
import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"strings"
	"syscall"
	"unsafe"
)

type ProcDetails struct {
	BundleID            string          `json:"bundleID,omitempty"`
	Args                []string        `json:"args,omitempty"`
	Path                string          `json:"path,omitempty"`
	User                string          `json:"user,omitempty"`
	FullUsername        string          `json:"full_username,omitempty"`
	Env                 json.RawMessage `json:"env,omitempty"`
	SandboxPath         string          `json:"sandboxcontainer,omitempty"`
	Pid                 int             `json:"pid,omitempty"`
	ScriptingProperties json.RawMessage `json:"scripting_properties,omitempty"`
	Name                string          `json:"name,omitempty"`
	Ppid                int             `json:"ppid"`
	Source              string          `json:"source"`
	ResponsiblePid      int             `json:"responsible_pid"`
	SubmittedPid        int             `json:"submitted_pid"`
	ParentPid           int             `json:"parent_pid"`
	BackingPlist        string          `json:"backing_plist"`
}

type DarwinProcess struct {
	pid                 int
	ppid                int
	binary              string
	architecture        string
	owner               string
	args                []string
	env                 map[string]string
	sandboxpath         string
	scriptingproperties map[string]interface{}
	name                string
	bundleid            string
	additionalInfo      map[string]interface{}
}

func (p *DarwinProcess) Pid() int {
	return p.pid
}

func (p *DarwinProcess) PPid() int {
	return p.ppid
}

func (p *DarwinProcess) Executable() string {
	return p.name
}

func (p *DarwinProcess) Arch() string {
	return p.architecture
}

func (p *DarwinProcess) Owner() string {
	return p.owner
}

func (p *DarwinProcess) BinPath() string {
	return p.binary
}

func (p *DarwinProcess) ProcessArguments() []string {
	return p.args
}

func (p *DarwinProcess) ProcessEnvironment() map[string]string {
	return p.env
}

func (p *DarwinProcess) SandboxPath() string {
	return p.sandboxpath
}

func (p *DarwinProcess) ScriptingProperties() map[string]interface{} {
	return p.scriptingproperties
}

func (p *DarwinProcess) Name() string {
	return p.name
}

func (p *DarwinProcess) BundleID() string {
	return p.bundleid
}

func (p *DarwinProcess) AdditionalInfo() map[string]interface{} {
	return map[string]interface{}{
		"ppid":            p.additionalInfo["ppid"],
		"source":          p.additionalInfo["source"],
		"responsible_pid": p.additionalInfo["responsible_pid"],
		"submitted_pid":   p.additionalInfo["submitted_pid"],
		"parent_pid":      p.additionalInfo["parent_pid"],
		"backing_plist":   p.additionalInfo["backing_plist"],
	}
}

func findProcess(pid int) (Process, error) {
	ps, err := Processes()
	if err != nil {
		return nil, err
	}

	for _, p := range ps {
		if p.Pid() == pid {
			return p, nil
		}
	}

	return nil, nil
}

func Processes() ([]Process, error) {
	buf, err := darwinSyscall()
	if err != nil {
		return nil, err
	}

	procs := make([]*kinfoProc, 0, 50)
	k := 0
	for i := _KINFO_STRUCT_SIZE; i < buf.Len(); i += _KINFO_STRUCT_SIZE {
		proc := &kinfoProc{}
		err = binary.Read(bytes.NewBuffer(buf.Bytes()[k:i]), binary.LittleEndian, proc)
		if err != nil {
			return nil, err
		}

		k = i
		procs = append(procs, proc)
	}

	darwinProcs := make([]Process, len(procs))
	for i, p := range procs {
		cpid := C.int(p.Pid)
		cresult := C.GetProcInfo(cpid)
		raw := C.GoString(cresult)
		r := []byte(raw)
		pinfo := ProcDetails{}
		var envJson map[string]string
		var scrptProps map[string]interface{}
		_ = json.Unmarshal(r, &pinfo)
		_ = json.Unmarshal(pinfo.Env, &envJson)
		// fixing an issue where some ENV values have nested JSON
		for key, _ := range envJson {
			envJson[key] = strings.ReplaceAll(envJson[key], `"`, `\\"`)
			envJson[key] = strings.ReplaceAll(envJson[key], "%22", "\\\"")
		}
		_ = json.Unmarshal(pinfo.ScriptingProperties, &scrptProps)
		darwinProcs[i] = &DarwinProcess{
			pid:                 int(p.Pid),
			ppid:                pinfo.Ppid,
			binary:              pinfo.Path,
			owner:               pinfo.User,
			args:                pinfo.Args,
			env:                 envJson,
			sandboxpath:         pinfo.SandboxPath,
			scriptingproperties: scrptProps,
			name:                pinfo.Name,
			bundleid:            pinfo.BundleID,
			additionalInfo: map[string]interface{}{
				"ppid":            pinfo.Ppid,
				"source":          pinfo.Source,
				"responsible_pid": pinfo.ResponsiblePid,
				"submitted_pid":   pinfo.SubmittedPid,
				"parent_pid":      pinfo.ParentPid,
				"backing_plist":   pinfo.BackingPlist,
			},
		}

	}

	return darwinProcs, nil
}

func darwinSyscall() (*bytes.Buffer, error) {
	mib := [4]int32{_CTRL_KERN, _KERN_PROC, _KERN_PROC_ALL, 0}
	size := uintptr(0)

	_, _, errno := syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		4,
		0,
		uintptr(unsafe.Pointer(&size)),
		0,
		0)

	if errno != 0 {
		return nil, errno
	}

	bs := make([]byte, size)
	_, _, errno = syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		4,
		uintptr(unsafe.Pointer(&bs[0])),
		uintptr(unsafe.Pointer(&size)),
		0,
		0)

	if errno != 0 {
		return nil, errno
	}

	return bytes.NewBuffer(bs[0:size]), nil
}

const (
	_CTRL_KERN         = 1
	_KERN_PROC         = 14
	_KERN_PROC_ALL     = 0
	_KINFO_STRUCT_SIZE = 648
)

type kinfoProc struct {
	_    [40]byte
	Pid  int32
	_    [199]byte
	Comm [16]byte
	_    [301]byte
	PPid int32
	_    [84]byte
}
