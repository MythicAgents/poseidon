#ifndef SHELL_MEMORY
#define SHELL_MEMORY

void* allocArgv(int);
void addArg(void*, char*, int);
int execMachO(char*, int, int, void*);
#endif
