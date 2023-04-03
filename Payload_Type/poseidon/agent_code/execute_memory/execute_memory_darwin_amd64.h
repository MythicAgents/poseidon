
#include <mach-o/dyld.h>
#include <sys/stat.h> 
#include <sys/mman.h> 
#include <fcntl.h>
#include <unistd.h>
#include <stdint.h>
#include <stdio.h>


extern char* executeMemory(void* memory, int memory_size, char* functionName, char* functionName2, char* argString);

