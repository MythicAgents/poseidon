
#include <mach-o/dyld.h>
#include <sys/stat.h> 
#include <sys/mman.h> 
#include <fcntl.h>
#include <unistd.h>
#include <stdint.h>
#include <stdio.h>


extern char* executeLibrary(char* filePath, char* functionName, int argc, char** argv);

