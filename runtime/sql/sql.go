package sql

// #cgo darwin amd64 CFLAGS: -I${SRCDIR}/deps/darwin_amd64
// #cgo darwin arm64 CFLAGS: -I${SRCDIR}/deps/darwin_arm64
// #cgo linux amd64 CFLAGS: -I${SRCDIR}/deps/linux_amd64
// #cgo windows amd64 CFLAGS: -I${SRCDIR}/deps/windows_amd64
// #include <stdlib.h>
// #include <librillsql.h>
// void*(*my_malloc)(size_t) = malloc;
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/rilldata/rill/runtime/pkg/sharedlibrary"
)

type Isolate struct {
	isolate *C.graal_isolate_t
	library sharedlibrary.Library
}

func NewIsolate() *Isolate {
	lib, err := sharedlibrary.OpenEmbed(libraryFS, libraryPath)
	if err != nil {
		panic(err)
	}

	graalCreateIsolate, err := lib.FindFunc("graal_create_isolate")
	if err != nil {
		panic(err)
	}

	var isolate *C.graal_isolate_t
	var thread *C.graal_isolatethread_t
	params := &C.graal_create_isolate_params_t{
		reserved_address_space_size: 1024 * 1024 * 500,
	}

	status, _, _ := graalCreateIsolate.Call(uintptr(unsafe.Pointer(params)), uintptr(unsafe.Pointer(&isolate)), uintptr(unsafe.Pointer(&thread)))
	if status != 0 {
		panic(fmt.Errorf("failed to create isolate"))
	}

	return &Isolate{
		library: lib,
		isolate: isolate,
	}
}

func (i *Isolate) Close() error {
	graalDetachAllThreadsAndTearDownIsolate, err := i.library.FindFunc("graal_detach_all_threads_and_tear_down_isolate")
	if err != nil {
		panic(err)
	}

	thread := i.attachThread()

	status, _, _ := graalDetachAllThreadsAndTearDownIsolate.Call(uintptr(unsafe.Pointer(thread)))
	if status != 0 {
		return fmt.Errorf("isolate teardown failed")
	}

	return i.library.Close()
}

func (i *Isolate) ConvertSQL(sql string, schema string) string {
	convertSql, err := i.library.FindFunc("convert_sql")
	if err != nil {
		panic(err)
	}

	thread := i.attachThread()

	cSql := C.CString(sql)
	defer C.free(unsafe.Pointer(cSql))

	cSchema := C.CString(schema)
	defer C.free(unsafe.Pointer(cSchema))

	res, _, _ := convertSql.Call(uintptr(unsafe.Pointer(thread)), uintptr(unsafe.Pointer(C.my_malloc)), uintptr(unsafe.Pointer(cSql)), uintptr(unsafe.Pointer(cSchema)))
	if res == 0 {
		panic(fmt.Errorf("call to convert_sql failed"))
	}

	goRes := C.GoString((*C.char)(unsafe.Pointer(res)))
	C.free(unsafe.Pointer(res))

	return goRes
}

func (i *Isolate) attachThread() *C.graal_isolatethread_t {
	graalGetCurrentThread, err := i.library.FindFunc("graal_get_current_thread")
	if err != nil {
		panic(err)
	}

	threadPtr, _, _ := graalGetCurrentThread.Call(uintptr(unsafe.Pointer(i.isolate)))
	if threadPtr != 0 {
		return (*C.graal_isolatethread_t)(unsafe.Pointer(threadPtr))
	}

	graalAttachThread, err := i.library.FindFunc("graal_attach_thread")
	if err != nil {
		panic(err)
	}

	var thread *C.graal_isolatethread_t
	status, _, _ := graalAttachThread.Call(uintptr(unsafe.Pointer(i.isolate)), uintptr(unsafe.Pointer(&thread)))
	if status != 0 {
		panic(fmt.Errorf("failed to attach thread"))
	}

	return thread
}
