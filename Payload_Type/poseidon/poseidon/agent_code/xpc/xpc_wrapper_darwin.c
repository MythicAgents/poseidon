#include <dispatch/dispatch.h>
#include <Block.h>
#include "xpc_wrapper_darwin.h"
#include <stdio.h>
#include <objc/objc.h>


struct xpc_global_data {
	uint64_t	a;
	uint64_t	xpc_flags;
	mach_port_t	task_bootstrap_port;  
#ifndef _64
	uint32_t	padding;
#endif
	xpc_object_t	xpc_bootstrap_pipe; 
};

#define OS_ALLOC_ONCE_KEY_MAX	100

struct _os_alloc_once_s {
	long once;
	void *ptr;
};
extern struct _os_alloc_once_s _os_alloc_once_table[];

xpc_type_t TYPE_ERROR = XPC_TYPE_ERROR;
xpc_type_t TYPE_ARRAY = XPC_TYPE_ARRAY;
xpc_type_t TYPE_DATA = XPC_TYPE_DATA;
xpc_type_t TYPE_DICT = XPC_TYPE_DICTIONARY;
xpc_type_t TYPE_INT64 = XPC_TYPE_INT64;
xpc_type_t TYPE_UINT64 = XPC_TYPE_UINT64;
xpc_type_t TYPE_STRING = XPC_TYPE_STRING;
xpc_type_t TYPE_UUID = XPC_TYPE_UUID;
xpc_type_t TYPE_BOOL = XPC_TYPE_BOOL;
xpc_type_t TYPE_DATE = XPC_TYPE_DATE;
xpc_type_t TYPE_FD = XPC_TYPE_FD;
xpc_type_t TYPE_CONNECTION = XPC_TYPE_CONNECTION;
xpc_type_t TYPE_NULL = XPC_TYPE_NULL;
xpc_type_t TYPE_SHMEM = XPC_TYPE_SHMEM;

xpc_object_t ERROR_CONNECTION_INVALID = (xpc_object_t) XPC_ERROR_CONNECTION_INVALID;
xpc_object_t ERROR_CONNECTION_INTERRUPTED = (xpc_object_t) XPC_ERROR_CONNECTION_INTERRUPTED;
xpc_object_t ERROR_CONNECTION_TERMINATED = (xpc_object_t) XPC_ERROR_TERMINATION_IMMINENT;

const ptr_to_uuid_t ptr_to_uuid(void *p) { return (ptr_to_uuid_t)p; }


xpc_object_t XpcLaunchdListServices(char *ServiceName) {
    // launchctl list [com.itsafeature.testing]
  xpc_object_t dict = xpc_dictionary_create(NULL, NULL, 0);
  xpc_dictionary_set_uint64(dict, "subsystem", 3);
  xpc_dictionary_set_uint64(dict, "handle",0);
  xpc_dictionary_set_uint64(dict, "routine", ROUTINE_LIST);
  xpc_dictionary_set_uint64(dict, "type",7); //used to be 1?

  if (ServiceName)
  {
    xpc_dictionary_set_string(dict, "name", ServiceName);
  }
  else
  {
    xpc_dictionary_set_bool(dict, "legacy", 1);
  }
  
  xpc_object_t *outDict = NULL;
  struct xpc_global_data  *xpc_gd  = (struct xpc_global_data *)  _os_alloc_once_table[1].ptr;

  int rc = xpc_pipe_routine (xpc_gd->xpc_bootstrap_pipe, dict, &outDict);
  
  if (outDict != NULL)
  {
    return outDict;
  }
  
  xpc_object_t errDict = xpc_dictionary_create(NULL, NULL, 0);
  xpc_dictionary_set_string(errDict, "error", "xpc_bootstrap_pipe returned a null dictionary");

  return errDict;
}

char* XpcLaunchdPrint(char *ServiceName) {

  vm_address_t addr;
  vm_size_t sz = 0x1400000;
  xpc_object_t shmem;
  const char *name = NULL;
  xpc_object_t dict = xpc_dictionary_create(NULL, NULL, 0);
  if(ServiceName){
        // launchctl print gui/501/com.itsafeature.testing
      int ret = launchctl_setup_xpc_dict_for_service_name(ServiceName, dict, &name);
      if(ret != 0){
        switch(ret){
            case ENODOMAIN:
                return "Error: No Domain";
            case ENOSERVICE:
                return "Error: No Service";
            case E2BIMPL:
                return "Error: 2BIMPL";
            case EUSAGE:
                return "Error: Usage";
            case EBADRESP:
                return "Error: Bad Response";
            case EDEPRECATED:
                return "Error: Deprecated";
            case EBADNAME:
                return "Error: Bad Name, needs to be in x/y/z format";
            default: {
                return "error";
            }

        }
      }
      xpc_dictionary_set_uint64(dict, "routine", ROUTINE_DUMP_PROCESS);
      xpc_dictionary_set_uint64(dict, "subsystem", ROUTINE_DUMP_PROCESS >> 8);
  }else{
        // launchctl print-cache
      xpc_dictionary_set_uint64(dict, "subsystem", 3);
      xpc_dictionary_set_uint64(dict, "handle",0);
      xpc_dictionary_set_uint64(dict, "routine", ROUTINE_PRINT);
      xpc_dictionary_set_uint64(dict, "type",1);
      xpc_dictionary_set_bool(dict, "cache", true);
  }
  vm_allocate(mach_task_self(), &addr, sz, 0xf0000003);
  shmem = xpc_shmem_create( (void*)addr, sz);
  xpc_dictionary_set_value(dict, "shmem", shmem);

  xpc_object_t *outDict = NULL;
  struct xpc_global_data  *xpc_gd  = (struct xpc_global_data *)  _os_alloc_once_table[1].ptr;

  int rc = xpc_pipe_routine (xpc_gd->xpc_bootstrap_pipe, dict, &outDict);

  if (outDict != NULL)
  {
    uint64_t written;
    written = xpc_dictionary_get_uint64(outDict, "bytes-written");
    if(written <= sz){
        if(written == 0){
            vm_deallocate(mach_task_self(), addr, sz);
            return "Wrote 0 bytes";
        } else {
            char* outputString = (char*)calloc(1, written + 1);
            strncpy(outputString, (char*)addr, written);
            vm_deallocate(mach_task_self(), addr, sz);
            return outputString;
        }
    }else{
        vm_deallocate(mach_task_self(), addr, sz);
        return "Wrote too many bytes";
    }
  }
  vm_deallocate(mach_task_self(), addr, sz);
  return "xpc_bootstrap_pipe returned a null dictionary";
}

char* XpcLaunchdDumpState(void) {
    // launchctl dumpstate
  vm_address_t addr;
  vm_size_t sz = 0x1400000;
  xpc_object_t shmem;
  xpc_object_t dict = xpc_dictionary_create(NULL, NULL, 0);
  xpc_dictionary_set_uint64(dict, "subsystem", 3);
  xpc_dictionary_set_uint64(dict, "handle",0);
  xpc_dictionary_set_uint64(dict, "routine", ROUTINE_DUMP_STATE);
  xpc_dictionary_set_uint64(dict, "type",1);

  vm_allocate(mach_task_self(), &addr, sz, 0xf0000003);
  shmem = xpc_shmem_create( (void*)addr, sz);
  xpc_dictionary_set_value(dict, "shmem", shmem);

  xpc_object_t *outDict = NULL;
  struct xpc_global_data  *xpc_gd  = (struct xpc_global_data *)  _os_alloc_once_table[1].ptr;

  int rc = xpc_pipe_routine (xpc_gd->xpc_bootstrap_pipe, dict, &outDict);

  if (outDict != NULL)
  {
    uint64_t written;
    written = xpc_dictionary_get_uint64(outDict, "bytes-written");
    if(written <= sz){
        if(written == 0){
            vm_deallocate(mach_task_self(), addr, sz);
            return "Wrote 0 bytes";
        } else {
            char* outputString = (char*)calloc(1, written + 1);
            strncpy(outputString, (char*)addr, written);
            vm_deallocate(mach_task_self(), addr, sz);
            return outputString;
        }
    }else{
        vm_deallocate(mach_task_self(), addr, sz);
        return "Wrote too many bytes";
    }
  }
  vm_deallocate(mach_task_self(), addr, sz);
  return "xpc_bootstrap_pipe returned a null dictionary";
}

xpc_object_t XpcLaunchdServiceControl(char *ServiceName, int StartStop) {
    // launchctl {start|stop} com.itsafeature.testing
  xpc_object_t dict = xpc_dictionary_create(NULL, NULL, 0);
  xpc_dictionary_set_uint64 (dict, "subsystem", 3);               
  xpc_dictionary_set_uint64(dict, "type",7); // was 1?
  xpc_dictionary_set_uint64(dict, "handle",0); 		  
  xpc_dictionary_set_string(dict, "name", ServiceName);                     
  xpc_dictionary_set_bool(dict, "legacy", 1);                       
  xpc_dictionary_set_uint64(dict, "routine", StartStop ? ROUTINE_START : ROUTINE_STOP);     

  xpc_object_t	*outDict = NULL;
  struct xpc_global_data  *xpc_gd  = (struct xpc_global_data *)  _os_alloc_once_table[1].ptr;

  int rc = xpc_pipe_routine (xpc_gd->xpc_bootstrap_pipe, dict, &outDict);

  if (outDict != NULL)
  {
    return outDict;
  }

  xpc_object_t errDict = xpc_dictionary_create(NULL, NULL, 0);
  xpc_dictionary_set_string(errDict, "error", "xpc_bootstrap_pipe returned a null dictionary");

  return errDict;
}

xpc_object_t XpcLaunchdServiceControlEnableDisable(char *ServiceName, int Enable) {
    // launchctl {enable|disable} gui/501/com.itsafeature.testing
  xpc_object_t dict = xpc_dictionary_create(NULL, NULL, 0);
  const char *name = NULL;
  int ret = launchctl_setup_xpc_dict_for_service_name(ServiceName, dict, &name);
  if (ret != 0){
    xpc_object_t errDict = xpc_dictionary_create(NULL, NULL, 0);
    xpc_dictionary_set_string(errDict, "error", "Failed to initialize service name, needs to be in x/y/z format");
    return errDict;
  }
  if (name == NULL) {
  	    xpc_object_t errDict = xpc_dictionary_create(NULL, NULL, 0);
        xpc_dictionary_set_string(errDict, "error", "Bad service name, needs to be in x/y/z format");
        return errDict;
  }
  xpc_dictionary_set_uint64 (dict, "subsystem", 3);
  //xpc_dictionary_set_uint64(dict, "type",8);
  //xpc_dictionary_set_uint64(dict, "handle",uid); //user id, typically 501
  //xpc_dictionary_set_string(dict, "name", ServiceName);
  xpc_object_t s = xpc_string_create(name);
  xpc_object_t names = xpc_array_create(&s, 1);
  xpc_dictionary_set_value(dict, "names", names);
  xpc_dictionary_set_uint64(dict, "routine", Enable ? ROUTINE_ENABLE : ROUTINE_DISABLE);

  xpc_object_t	*outDict = NULL;
  struct xpc_global_data  *xpc_gd  = (struct xpc_global_data *)  _os_alloc_once_table[1].ptr;

  int rc = xpc_pipe_routine (xpc_gd->xpc_bootstrap_pipe, dict, &outDict);

  if (outDict != NULL)
  {
    return outDict;
  }

  xpc_object_t errDict = xpc_dictionary_create(NULL, NULL, 0);
  xpc_dictionary_set_string(errDict, "error", "xpc_bootstrap_pipe returned a null dictionary");

  return errDict;
}

xpc_object_t XpcLaunchdSubmitJob(char *Program, char *ServiceName, int KeepAlive) {
  xpc_object_t dict = xpc_dictionary_create(NULL, NULL, 0);
  xpc_object_t request = xpc_dictionary_create(NULL, NULL, 0);
  xpc_object_t submitJob = xpc_dictionary_create(NULL, NULL, 0);

  xpc_dictionary_set_bool(submitJob, "KeepAlive", KeepAlive);
  xpc_dictionary_set_string (submitJob, "Program", Program);
  xpc_dictionary_set_string (submitJob, "Label", ServiceName);

  xpc_object_t programArguments = xpc_array_create(NULL, 0);
  xpc_dictionary_set_value(submitJob, "ProgramArguments", programArguments);

  xpc_dictionary_set_value (request, "SubmitJob", submitJob);
  xpc_dictionary_set_value (dict, "request", request);
  xpc_dictionary_set_uint64 (dict, "subsystem", 7);              
  xpc_dictionary_set_uint64(dict, "type",7);                      
  xpc_dictionary_set_uint64(dict, "handle",0); 		  
  xpc_dictionary_set_uint64(dict, "routine", ROUTINE_SUBMIT);      

  xpc_object_t	*outDict = NULL;

  struct xpc_global_data  *xpc_gd  = (struct xpc_global_data *)  _os_alloc_once_table[1].ptr;

  int rc = xpc_pipe_routine(xpc_gd->xpc_bootstrap_pipe, dict, &outDict);
   
  if (outDict != NULL)
  {
    return outDict;
  }



  xpc_object_t errDict = xpc_dictionary_create(NULL, NULL, 0);
  xpc_dictionary_set_string(errDict, "error", "xpc_bootstrap_pipe returned a null dictionary");

  return errDict;
}

xpc_object_t XpcLaunchdRemove(char *ServiceName) {
  xpc_object_t dict = xpc_dictionary_create(NULL, NULL,0);
  xpc_dictionary_set_uint64 (dict, "subsystem", 3);
  xpc_dictionary_set_uint64(dict, "type",7);
  xpc_dictionary_set_uint64(dict, "handle",0);
  xpc_dictionary_set_uint64(dict, "routine", ROUTINE_REMOVE) ;
  xpc_dictionary_set_string (dict,"name", ServiceName);
  xpc_dictionary_set_bool(dict, "legacy", true);

  xpc_object_t	*outDict = NULL;
  struct xpc_global_data  *xpc_gd  = (struct xpc_global_data *)  _os_alloc_once_table[1].ptr;

  int rc = xpc_pipe_routine (xpc_gd->xpc_bootstrap_pipe, dict, &outDict);

  if (outDict != NULL)
  {
    return outDict;
  }
  // if we get here there was a problem
  xpc_object_t errDict = xpc_dictionary_create(NULL, NULL, 0);
  xpc_dictionary_set_string(errDict, "error", "xpc_bootstrap_pipe returned a null dictionary");

  return errDict;
}

xpc_object_t XpcLaunchdAsUser(char *program, int uid) {
  xpc_object_t dict = xpc_dictionary_create(NULL, NULL,0);
  xpc_dictionary_set_uint64 (dict, "subsystem", 3);
  xpc_dictionary_set_uint64(dict, "type",1);
  xpc_dictionary_set_uint64(dict, "handle",0);
  xpc_dictionary_set_uint64(dict, "routine", 835);
  xpc_dictionary_set_uint64(dict, "uid",uid);

  xpc_object_t	*outDict = NULL;
  struct xpc_global_data  *xpc_gd  = (struct xpc_global_data *)  _os_alloc_once_table[1].ptr;

  int rc = xpc_pipe_routine (xpc_gd->xpc_bootstrap_pipe, dict, &outDict);

  if (outDict != NULL)
  {
    return outDict;
  }
  // if we get here there was a problem
  xpc_object_t errDict = xpc_dictionary_create(NULL, NULL, 0);
  xpc_dictionary_set_string(errDict, "error", "xpc_bootstrap_pipe returned a null dictionary");

  return errDict;
}

xpc_object_t XpcLaunchdGetManagerUID() {
    // launchctl manageruid
   xpc_object_t dict = xpc_dictionary_create(NULL, NULL,0);
   xpc_dictionary_set_uint64 (dict, "subsystem", 6);
   xpc_dictionary_set_uint64(dict, "type",7);
   xpc_dictionary_set_uint64(dict, "handle",0);
   xpc_dictionary_set_uint64(dict, "routine", 301) ;
   xpc_dictionary_set_uint64(dict, "outgsk",3);
   xpc_dictionary_set_bool(dict, "get", true);
    xpc_dictionary_set_bool(dict, "self", true);
   xpc_object_t	*outDict = NULL;
   struct xpc_global_data  *xpc_gd  = (struct xpc_global_data *)  _os_alloc_once_table[1].ptr;

   int rc = xpc_pipe_routine (xpc_gd->xpc_bootstrap_pipe, dict, &outDict);

   if (outDict != NULL)
   {
     return outDict;
   }


   // if we get here there was a problem
   xpc_object_t errDict = xpc_dictionary_create(NULL, NULL, 0);
   xpc_dictionary_set_string(errDict, "error", "xpc_bootstrap_pipe returned a null dictionary");

   return errDict;
 }

char* XpcLaunchdGetProcInfo(unsigned long pid) {
  char *pointer;
  int ret;
  pointer = tmpnam(NULL);
  int fd = open(pointer, O_WRONLY | O_CREAT |  O_TRUNC,
    S_IRUSR | S_IWUSR | S_IRGRP | S_IROTH);
  //int fd = fileno(tmp);
  xpc_object_t dict = xpc_dictionary_create(NULL, NULL, 0);
  xpc_dictionary_set_uint64(dict, "routine",ROUTINE_DUMP_PROCESS); 
  xpc_dictionary_set_uint64 (dict, "subsystem", 2); 
  xpc_dictionary_set_fd(dict, "fd",fd);                             
  xpc_dictionary_set_int64(dict, "pid",pid);                       
  xpc_object_t	*outDict = NULL;

  struct xpc_global_data  *xpc_gd  = (struct xpc_global_data *)  _os_alloc_once_table[1].ptr;

  int rc = xpc_pipe_routine (xpc_gd->xpc_bootstrap_pipe, dict, &outDict);
  //close(fd);
  return pointer;
}

xpc_object_t XpcLaunchdLoadPlist(char *Plist, int Legacy) {
  xpc_object_t dict = xpc_dictionary_create(NULL, NULL, 0);
  xpc_object_t s = xpc_string_create(Plist);
  xpc_object_t paths = xpc_array_create(&s, 1);
  xpc_dictionary_set_value(dict, "paths", paths);

  xpc_dictionary_set_uint64 (dict, "subsystem", 3);              
  xpc_dictionary_set_bool(dict, "enable", true); // launchctl load -w (sets this to true), (no -w this is false)
	if (Legacy) xpc_dictionary_set_bool(dict, "legacy", true);
  xpc_dictionary_set_bool(dict, "legacy-load", true);
  xpc_dictionary_set_uint64(dict, "type",7);                     
  xpc_dictionary_set_uint64(dict, "handle",0); 		   
  xpc_dictionary_set_string(dict, "session", "Aqua");
  xpc_dictionary_set_uint64(dict, "routine", ROUTINE_LOAD);

  xpc_object_t	*outDict = NULL;
  struct xpc_global_data  *xpc_gd  = (struct xpc_global_data *)  _os_alloc_once_table[1].ptr;


  int rc = xpc_pipe_routine(xpc_gd->xpc_bootstrap_pipe, dict, &outDict);
  if (rc == 0) {
    int err = xpc_dictionary_get_int64 (outDict, "error");
    if (err) printf("Error:  %d - %s\n", err, xpc_strerror(err));
    return outDict;
  }

  if (outDict != NULL)
  {
    return outDict;
  }


 
  xpc_object_t errDict = xpc_dictionary_create(NULL, NULL, 0);
  xpc_dictionary_set_string(errDict, "error", "xpc_bootstrap_pipe returned a null dictionary");

  return errDict;
}

xpc_object_t XpcLaunchdUnloadPlist(char *Plist) {
  xpc_object_t dict = xpc_dictionary_create(NULL, NULL, 0);
  xpc_object_t s = xpc_string_create(Plist);
  xpc_object_t paths = xpc_array_create(&s, 1);
  xpc_dictionary_set_value (dict, "paths", paths);

  xpc_dictionary_set_uint64 (dict, "subsystem", 3);             
  xpc_dictionary_set_bool(dict, "disable", true); // launchctl unload -w (sets this to true), (no -w this is false)
  xpc_dictionary_set_bool(dict, "legacy-load", true);
  xpc_dictionary_set_bool(dict, "legacy", true);
  xpc_dictionary_set_uint64(dict, "type",7);                      
  xpc_dictionary_set_uint64(dict, "handle",0); 		   
  xpc_dictionary_set_string(dict, "session", "Aqua");
  xpc_dictionary_set_uint64(dict, "routine", ROUTINE_UNLOAD);
  xpc_dictionary_set_bool(dict, "no-einprogress", true); // seen on ventura


  xpc_object_t	*outDict = NULL;
  struct xpc_global_data  *xpc_gd  = (struct xpc_global_data *)  _os_alloc_once_table[1].ptr;


  int rc = xpc_pipe_routine (xpc_gd->xpc_bootstrap_pipe, dict, &outDict);
  if (rc == 0) {
    int err = xpc_dictionary_get_int64 (outDict, "error");
    if (err) printf("Error:  %d - %s\n", err, xpc_strerror(err));
    return outDict;
  }

  if (outDict != NULL)
  {
    return outDict;
  }


 
  xpc_object_t errDict = xpc_dictionary_create(NULL, NULL, 0);
  xpc_dictionary_set_string(errDict, "error", "xpc_bootstrap_pipe returned a null dictionary");

  return errDict;
}

xpc_connection_t XpcConnect(char *service, uintptr_t ctx, int privileged) {
    dispatch_queue_t queue = dispatch_queue_create(service, 0);
    xpc_connection_t conn = xpc_connection_create_mach_service(service, queue, privileged ? (uint64_t)XPC_CONNECTION_MACH_SERVICE_PRIVILEGED: 0);

    

    xpc_connection_set_event_handler(conn,
        Block_copy(^(xpc_object_t event) {
            handleXpcEvent(event, ctx); 
        })
    );

    xpc_connection_resume(conn);
    return conn;
}

void XpcSendMessage(xpc_connection_t conn, xpc_object_t message, bool release, bool reportDelivery) {
    xpc_connection_send_message(conn,  message);
    xpc_connection_send_barrier(conn, ^{
        
        if (reportDelivery) { 
            puts("message delivered");
        }
    });
    if (release) {
        xpc_release(message);
    }
}

void XpcArrayApply(uintptr_t v, xpc_object_t arr) {
  xpc_array_apply(arr, ^bool(size_t index, xpc_object_t value) {
    arraySet(v, index, value);
    return true;
  });
}

void XpcDictApply(uintptr_t v, xpc_object_t dict) {
  xpc_dictionary_apply(dict, ^bool(const char *key, xpc_object_t value) {
    dictSet(v, (char *)key, value);
    return true;
  });
}

void XpcUUIDGetBytes(void *v, xpc_object_t uuid) {
   const uint8_t *src = xpc_uuid_get_bytes(uuid);
   uint8_t *dest = (uint8_t *)v;

   for (int i=0; i < sizeof(uuid_t); i++) {
     dest[i] = src[i];
   }
}

// https://github.com/ProcursusTeam/launchctl/blob/main/xpc_helper.c#L148C1-L221C2
// modified to add support for gui/ service name
int
launchctl_setup_xpc_dict_for_service_name(char *servicetarget, xpc_object_t dict, const char **name)
{
	long handle = 0;

	if (name != NULL) {
		*name = NULL;
	}

	const char *split[3] = {NULL, NULL, NULL};
	for (int i = 0; i < 3; i++) {
		char *var = strsep(&servicetarget, "/");
		if (var == NULL)
			break;
		split[i] = var;
	}
	if (split[0] == NULL || split[0][0] == '\0')
		return EBADNAME;

	if (strcmp(split[0], "system") == 0) {
		xpc_dictionary_set_uint64(dict, "type", 1);
		xpc_dictionary_set_uint64(dict, "handle", 0);
		if (split[1] != NULL && split[1][0] != '\0') {
			xpc_dictionary_set_string(dict, "name", split[1]);
			if (name != NULL) {
				*name = split[1];
			}
		}
		return 0;
	} else if (strcmp(split[0], "user") == 0) {
		xpc_dictionary_set_uint64(dict, "type", 2);
	} else if (strcmp(split[0], "session") == 0) {
		xpc_dictionary_set_uint64(dict, "type", 4);
	} else if (strcmp(split[0], "pid") == 0) {
		xpc_dictionary_set_uint64(dict, "type", 5);
    } else if(strcmp(split[0], "gui") == 0) {
		xpc_dictionary_set_uint64(dict, "type", 8);
	} else {
		xpc_dictionary_set_uint64(dict, "type", 9);
	}
	if (split[1] != NULL) {
		if (handle == 0) {
			handle = strtol(split[1], NULL, 10);
			if (handle == -1)
				return EUSAGE;
		}
		xpc_dictionary_set_uint64(dict, "handle", handle);
		if (split[2] != NULL && split[2][0] != '\0') {
			xpc_dictionary_set_string(dict, "name", split[2]);
			if (name != NULL) {
				*name = split[2];
			}
		}
		return 0;
	}
	return EBADNAME;
}