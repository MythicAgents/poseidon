#ifndef LIBINJECT_DARWIN_ARM64_H
#define LIBINJECT_DARWIN_ARM64_H

#import <EndpointSecurity/EndpointSecurity.h>
#import <Security/Security.h>
#include <dlfcn.h>
#include <errno.h>
#include <libproc.h>
#include <mach/arm/thread_status.h>
#include <mach/error.h>
#include <mach/mach.h>
#include <mach/mach_vm.h>
#include <mach/thread_act.h>
#include <pthread.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <sys/sysctl.h>
#include <sys/types.h>
#include <unistd.h>

extern int inject(pid_t pid, char *lib);

#endif // LIBINJECT_DARWIN_ARM64_H
