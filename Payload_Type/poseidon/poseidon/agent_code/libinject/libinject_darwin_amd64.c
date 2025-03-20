#include "libinject_darwin_amd64.h"

#define STACK_SIZE 65536
#define CODE_SIZE 128
// 2025-03-19 T20:20:01 UTC: Changed to avoid reuse of existing dylib path
char shellcode_template[] =
"\x55"                            // push       rbp
"\x48\x89\xE5"                    // mov        rbp, rsp
"\x48\x83\xEC\x10"                // sub        rsp, 0x10
"\x48\x8D\x7D\xF8"                // lea        rdi, qword [rbp+var_8]
"\x31\xC0"                        // xor        eax, eax
"\x89\xC1"                        // mov        ecx, eax
"\x48\x8D\x15\x1A\x00\x00\x00"    // lea        rdx, qword ptr [rip + 0x1A]
"\x48\x89\xCE"                    // mov        rsi, rcx
"\x48\xB8"                        // movabs     rax, pthread_create_from_mach_thread
"PTHRDCRT"
"\xFF\xD0"                        // call       rax
"\x89\x45\xF4"                    // mov        dword [rbp+var_C], eax
"\x48\x83\xC4\x10"                // add        rsp, 0x10
"\x5D"                            // pop        rbp
"\xEB\xFE"                        // jmp        0x0
"\xC3"                            // ret
"\x55"                            // push       rbp
"\x48\x89\xE5"                    // mov        rbp, rsp
"\x48\x83\xEC\x10"                // sub        rsp, 0x10
"\xBE\x01\x00\x00\x00"            // mov        esi, 0x1
"\x48\x89\x7D\xF8"                // mov        qword [rbp+var_8], rdi
"\x48\x8D\x3D\x1D\x00\x00\x00"    // lea        rdi, qword ptr [rip + 0x2c]
"\x48\xB8"                        // movabs     rax, dlopen
"DLOPEN__"
"\xFF\xD0"                        // call       rax
"\x31\xF6"                        // xor        esi, esi
"\x89\xF7"                        // mov        edi, esi
"\x48\x89\x45\xF0"                // mov        qword [rbp+var_10], rax
"\x48\x89\xF8"                    // mov        rax, rdi
"\x48\x83\xC4\x10"                // add        rsp, 0x10
"\x5D"                            // pop        rbp
"\xC3"                            // ret
"LIBLIBLIBLIB"
"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"
"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"
"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"
"\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00";
task_t taskForPidWrapper(pid_t pid)
{
    host_t  myhost = mach_host_self();
    host_t  host_priv;
    mach_port_t psDefault;
    mach_port_t psDefault_control;
    task_array_t  tasks;
    mach_msg_type_number_t numTasks;
    int i;
    thread_array_t   threads;
    thread_info_data_t tInfo;
    kern_return_t kr;
    host_get_host_priv_port(mach_host_self(), &host_priv);
    kr = processor_set_default(host_priv, &psDefault);
    processor_set_name_array_t  *psets = malloc(1024);
    mach_msg_type_number_t    psetCount;
    kr = host_processor_sets (host_priv, psets, &psetCount);
    kr = host_processor_set_priv(host_priv, psDefault, &psDefault_control);
    if (kr != KERN_SUCCESS) { fprintf(stderr, "host_processor_set_priv failed with error %x\n", kr);  mach_error("host_processor_set_priv",kr); return 0;}
    numTasks=1000;
    kr = processor_set_tasks(psDefault_control, &tasks, &numTasks);
    if (kr != KERN_SUCCESS) { fprintf(stderr,"processor_set_tasks failed with error %x\n",kr);return 0; }
    for (i = 0; i < numTasks; i++)
    {
        char name[128];
        int p;
        pid_for_task(tasks[i], &p);
        //int rc=  proc_name(pid, name, 128);
        if (p == pid)
        {
            return tasks[i];
        }
    }
    return 0;
}

int inject(pid_t pid, char* lib)
{
    task_t remoteTask;
    struct stat buf;
    int rc = stat (lib, &buf);
    if (rc != 0)
    {
        return (-9);
    }
    mach_error_t kr = 0;
    remoteTask = taskForPidWrapper(pid);
    if (remoteTask == NULL)
    {
        return (-2);
    }
    mach_vm_address_t remoteStack64 = (vm_address_t) NULL;
    mach_vm_address_t remoteCode64 = (vm_address_t) NULL;
    kr = mach_vm_allocate( remoteTask, &remoteStack64, STACK_SIZE, VM_FLAGS_ANYWHERE);
    if (kr != KERN_SUCCESS)
	{
		return (-2);
	}
    remoteCode64 = (vm_address_t) NULL;
    kr = mach_vm_allocate( remoteTask, &remoteCode64, CODE_SIZE, VM_FLAGS_ANYWHERE );
    if (kr != KERN_SUCCESS)
    {
        return (-2);
    }

    // 2025-03-19 T20:20:01 UTC: Updated to avoid reuse of dylab paths
    size_t shellcode_size = sizeof(shellcode_template);
    char *injectedCode = malloc(shellcode_size);                    // Allocate new injectedCode byte array
    if (!injectedCode) {
        perror("[-] Failed to allocate memory for shellcode\n");
        return -1;
    }
    memcpy(injectedCode, shellcode_template, shellcode_size);       // Copy fresh shellcode for patching

    int i = 0;
    char *possiblePatchLocation = (injectedCode );
    for (i = 0 ; i < 0x100; i++)
    {
        extern void *_pthread_set_self;
        possiblePatchLocation++;
        uint64_t addrOfPthreadCreateFromMachThread = (uint64_t)dlsym( RTLD_DEFAULT, "pthread_create_from_mach_thread"); //(uint64_t) _pthread_set_self; Get the address, because we pull the address from the sharedcache.
        uint64_t addrOfDlopen = (uint64_t) dlopen;
        uint64_t addrOfSleep = (uint64_t) sleep;
        if (memcmp (possiblePatchLocation, "PTHRDCRT", 8) == 0)
        {
            memcpy(possiblePatchLocation, &addrOfPthreadCreateFromMachThread,8);
        }
        if (memcmp(possiblePatchLocation, "DLOPEN__", 6) == 0)
        {
            memcpy(possiblePatchLocation, &addrOfDlopen, sizeof(uint64_t));
        }
        if (memcmp(possiblePatchLocation, "LIBLIBLIB", 9) == 0)
        {
            strcpy(possiblePatchLocation, lib );
        }
    }
	kr = mach_vm_write(remoteTask,remoteCode64, (vm_address_t) injectedCode, 0xa9);                       // Length of the source
    if (kr != KERN_SUCCESS)
	{
		return (-3);
	}
    kr  = vm_protect(remoteTask, remoteCode64, 0x70, FALSE, VM_PROT_READ | VM_PROT_EXECUTE);
    kr  = vm_protect(remoteTask, remoteStack64, STACK_SIZE, TRUE, VM_PROT_READ | VM_PROT_WRITE);
    if (kr != KERN_SUCCESS)
	{
		return (-4);
	}
    x86_thread_state64_t remoteThreadState64;
    thread_act_t         remoteThread;
    memset(&remoteThreadState64, '\0', sizeof(remoteThreadState64) );
    remoteStack64 += (STACK_SIZE / 2);
    const char* p = (const char*) remoteCode64;
    remoteThreadState64.__rip = (u_int64_t) (vm_address_t) remoteCode64;
    remoteThreadState64.__rsp = (u_int64_t) remoteStack64;
    remoteThreadState64.__rbp = (u_int64_t) remoteStack64;
    kr = thread_create_running( remoteTask, x86_THREAD_STATE64, (thread_state_t) &remoteThreadState64, x86_THREAD_STATE64_COUNT, &remoteThread );
    if (kr != KERN_SUCCESS)
    {
		return (-3);
    }
    return (0);
}
EOF