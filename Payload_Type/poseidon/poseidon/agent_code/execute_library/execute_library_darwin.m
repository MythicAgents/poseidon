//go:build darwin && (execute_library || debug)

#include "execute_library_darwin.h"
#import <Foundation/Foundation.h>
#include "stdio.h"
#include <dlfcn.h>
#include <string.h>
#include <stdlib.h>

char* executeLibrary(char* filePath, char* functionName, int argc, char** argv){
    void* handle = dlopen(filePath, RTLD_NOW);
    if(handle != 0){
        char*(*function)(int c, char** argv) = dlsym(handle, functionName);
        if(function != 0){
            char* output = function(argc, argv);
            char* copiedOutput = output != NULL ? strdup(output) : strdup("");
            dlclose(handle);
            return copiedOutput;
        }
        dlclose(handle);
        return strdup("Failed to find function");
    }
    return strdup("Failed to open dynamic library");
}
