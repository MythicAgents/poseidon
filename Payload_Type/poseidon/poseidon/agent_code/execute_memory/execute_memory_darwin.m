#include "execute_memory_darwin.h"
#import <Foundation/Foundation.h>
#include "stdio.h"

char* executeMemory(void* memory, int memory_size, char* functionName, char* functionName2, char* argString){

	NSObjectFileImage fileImage = NULL;
	NSModule module = NULL;
	NSSymbol symbol = NULL;
	int pid = 0;
	//struct stat stat_buf;
	void* (*function)();
	//printf("setting memory value\n");
	if(memory_size < 20){
		return "Supplied file is too small";
	}
        *((uint8_t*)memory + 12) =  0x08;
        //printf("set value\n");
	NSCreateObjectFileImageFromMemory(memory, memory_size, &fileImage);
	//printf("created file image\n");
	if(fileImage == NULL){
		return "Failed to get File Image from memory";
	}
	module = NSLinkModule(fileImage, "module", NSLINKMODULE_OPTION_NONE);
	//printf("created module, %p\n", module);
	if(module == NULL){
		return "Failed to get module from memory image";
	}
	symbol = NSLookupSymbolInModule(module, functionName);
	//printf("got symbol, %p\n", symbol);
	if(symbol == NULL){
	    symbol = NSLookupSymbolInModule(module, functionName2);
	    //printf("got symbol, %p\n", symbol);
	    if(symbol == NULL){
	        return "Failed to find function name in module";
	    }
	}
	function = (void*(*)()) NSAddressOfSymbol(symbol);
	//printf("got function\n");
	if(function == NULL){
		return "Failed to find address of function";
	}
	char* output = NULL;
	@try{
	    output = (char*) function(argString);
	}@catch(NSException *e){
	    output = e.reason.UTF8String;
	}
	NSUnLinkModule(module, NSUNLINKMODULE_OPTION_NONE);
	NSDestroyObjectFileImage(fileImage);
	return output;
}
