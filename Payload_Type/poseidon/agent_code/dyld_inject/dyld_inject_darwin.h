#ifndef spawn_libinject_h
#define spawn_libinject_h

extern int dyld_inject(char* app, char* dylib, int hide);

#endif /* spawn_libinject_h */