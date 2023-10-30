#include "execute_library_darwin.h"
#import <Foundation/Foundation.h>
#include "stdio.h"
#include <dlfcn.h>

char* executeLibrary(char* filePath, char* functionName, int argc, char** argv){
    void* handle = dlopen(filePath, RTLD_NOW);
    if(handle){
        char*(*function)(int c, char** argv) = dlsym(handle, functionName);
        if(function){
            char* output = function(argc, argv);
            dlclose(handle);
            return output;
        }
        dlclose(handle);
        return "Failed to find function";
    }
    return "Failed to open dynamic library";
}
