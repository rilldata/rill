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
	b64 "encoding/base64"
	"fmt"
	"unsafe"

	"github.com/rilldata/rill/runtime/pkg/sharedlibrary"
	"github.com/rilldata/rill/runtime/sql/rpc"
	"google.golang.org/protobuf/proto"
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

func (i *Isolate) request(request *rpc.Request) *rpc.Response {
	f, err := i.library.FindFunc("request")
	if err != nil {
		panic(err)
	}

	thread := i.attachThread()

	bytes, _ := proto.Marshal(request)
	b64request := b64.StdEncoding.EncodeToString(bytes)

	cBytes := C.CString(b64request)
	defer C.free(unsafe.Pointer(cBytes))
	res, _, _ := f.Call(
		uintptr(unsafe.Pointer(thread)),
		uintptr(unsafe.Pointer(C.my_malloc)),
		uintptr(unsafe.Pointer(cBytes)),
	)
	if res == 0 {
		panic(fmt.Errorf("call to request failed"))
	}

	b64response := C.GoString((*C.char)(unsafe.Pointer(res)))
	C.free(unsafe.Pointer(res))

	var response rpc.Response
	decodedResponse, _ := b64.StdEncoding.DecodeString(b64response)
	proto.Unmarshal(decodedResponse, &response)

	return &response
}

func (i *Isolate) ConvertSQL(sql string, schema string, dialect string) string {
	convertSql, err := i.library.FindFunc("convert_sql")
	if err != nil {
		panic(err)
	}

	thread := i.attachThread()

	cSql := C.CString(sql)
	defer C.free(unsafe.Pointer(cSql))

	cSchema := C.CString(schema)
	defer C.free(unsafe.Pointer(cSchema))

	cDialect := C.CString(dialect)
	defer C.free(unsafe.Pointer(cDialect))

	res, _, _ := convertSql.Call(
		uintptr(unsafe.Pointer(thread)),
		uintptr(unsafe.Pointer(C.my_malloc)),
		uintptr(unsafe.Pointer(cSql)),
		uintptr(unsafe.Pointer(cSchema)),
		uintptr(unsafe.Pointer(cDialect)),
	)
	if res == 0 {
		panic(fmt.Errorf("call to convert_sql failed"))
	}

	goRes := C.GoString((*C.char)(unsafe.Pointer(res)))
	C.free(unsafe.Pointer(res))

	return goRes
}

func (i *Isolate) getAST(sql string, schema string) []byte {
	getAST, err := i.library.FindFunc("get_ast")
	if err != nil {
		panic(err)
	}

	thread := i.attachThread()

	cSql := C.CString(sql)
	defer C.free(unsafe.Pointer(cSql))

	cSchema := C.CString(schema)
	defer C.free(unsafe.Pointer(cSchema))

	res, _, _ := getAST.Call(
		uintptr(unsafe.Pointer(thread)),
		uintptr(unsafe.Pointer(C.my_malloc)),
		uintptr(unsafe.Pointer(cSql)),
		uintptr(unsafe.Pointer(cSchema)),
	)
	if res == 0 {
		panic(fmt.Errorf("call to get_ast failed"))
	}

	goResString := C.GoString((*C.char)(unsafe.Pointer(res)))
	goRes := []byte(goResString)
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
