package sql

// #cgo darwin amd64 CFLAGS: -I${SRCDIR}/deps/darwin_amd64
// #cgo darwin arm64 CFLAGS: -I${SRCDIR}/deps/darwin_arm64
// #cgo linux amd64 CFLAGS: -I${SRCDIR}/deps/linux_amd64
// #cgo windows amd64 CFLAGS: -I${SRCDIR}/deps/windows_amd64
// #include <stdlib.h>
// #include <librillsql.h>
// void*(*malloc_wrap)(size_t) = malloc;
import "C"

import (
	"encoding/base64"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"github.com/rilldata/rill/runtime/pkg/sharedlibrary"
	"github.com/rilldata/rill/runtime/sql/rpc"
	"google.golang.org/protobuf/proto"
)

var ErrIsolateClosed = errors.New("sql: isolate closed")

// Isolate represents a SQL library (GraalVM) isolate that we can interface with using the protobuf RPCs.
// Isolates have thread-local state, and since Go can arbitrarily create new threads, we handle the entire
// isolate lifecycle in a single thread-locked event loop (see eventLoop).
type Isolate struct {
	queue       chan *job
	closed      bool
	closedMu    sync.RWMutex
	closeDoneCh chan struct{}
	closeErr    error
}

// Job represents one request that's channeled into the event loop through the isolate's queue channel.
// Jobs can be submitted from any thread using Request.
type job struct {
	req    *rpc.Request
	resp   *rpc.Response
	doneCh chan struct{}
}

// OpenIsolate creates an isolate and starts its event loop
func OpenIsolate() *Isolate {
	i := &Isolate{
		queue:       make(chan *job),
		closeDoneCh: make(chan struct{}),
	}

	go i.eventLoop()

	return i
}

// Close will finish any outstanding requests, then teardown the GraalVM isolate and end the event loop.
func (i *Isolate) Close() error {
	i.closedMu.Lock()
	i.closed = true
	i.closedMu.Unlock()
	close(i.queue)
	<-i.closeDoneCh
	return i.closeErr
}

// Request makes thread-safe requests to the SQL library.
// It only errors if the isolate was closed. It's the caller's
// responsibility to handle response.Error.
func (i *Isolate) Request(req *rpc.Request) (*rpc.Response, error) {
	// Create job
	j := &job{
		req:    req,
		doneCh: make(chan struct{}),
	}

	// Safely add it to queue by checking the isolate hasn't been closed
	var err error
	i.closedMu.RLock()
	if !i.closed {
		i.queue <- j
	} else {
		err = ErrIsolateClosed
		close(j.doneCh)
	}
	i.closedMu.RUnlock()
	if err != nil {
		return nil, err
	}

	// Wait for the event loop to finish the job
	<-j.doneCh

	return j.resp, nil
}

// eventLoop handles the entire isolate lifecycle in a single function, which is locked
// to a single thread using Go's LockOSThread. This makes isolates simpler to reason about.
// It can accept jobs from any goroutine/thread over the i.queue channel.
func (i *Isolate) eventLoop() {
	// Lock everything in this function to a single thread
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Get funcs from SQL lib
	lib := getLibSQL()
	graalCreateIsolateFn, err := lib.FindFunc("graal_create_isolate")
	if err != nil {
		panic(err)
	}
	graalDetachAllThreadsAndTearDownIsolateFn, err := lib.FindFunc("graal_detach_all_threads_and_tear_down_isolate")
	if err != nil {
		panic(err)
	}
	requestFn, err := lib.FindFunc("request")
	if err != nil {
		panic(err)
	}

	// Create isolate
	var isolate *C.graal_isolate_t
	var thread *C.graal_isolatethread_t
	params := &C.graal_create_isolate_params_t{
		reserved_address_space_size: 1024 * 1024 * 500,
	}
	status, _, err := graalCreateIsolateFn.Call(uintptr(unsafe.Pointer(params)), uintptr(unsafe.Pointer(&isolate)), uintptr(unsafe.Pointer(&thread)))
	if status != 0 || err != nil {
		panic(fmt.Errorf("failed to create isolate"))
	}

	// Run main event loop, which receives jobs (requests) into this goroutine (which is OS thread locked) and processes them
	for {
		j, ok := <-i.queue

		// If the queue was closed, stop the event loop
		if !ok {
			close(i.closeDoneCh)
			break
		}

		// Serialize request
		reqMsg, err := proto.Marshal(j.req)
		if err != nil {
			panic(fmt.Errorf("could not serialize request"))
		}
		reqMsgC := C.CString(base64.StdEncoding.EncodeToString(reqMsg))
		defer C.free(unsafe.Pointer(reqMsgC))

		// Call lib
		res, _, err := requestFn.Call(
			uintptr(unsafe.Pointer(thread)),
			uintptr(unsafe.Pointer(C.malloc_wrap)),
			uintptr(unsafe.Pointer(reqMsgC)),
		)
		if err != nil {
			panic(fmt.Errorf("failed to call request"))
		}
		if res == 0 {
			panic(fmt.Errorf("sql library returned null response"))
		}

		// Deserialize response
		resMsg64 := C.GoString((*C.char)(unsafe.Pointer(res)))
		C.free(unsafe.Pointer(res)) // SQL library mallocs result using the passed-in malloc_wrap
		resMsg, err := base64.StdEncoding.DecodeString(resMsg64)
		if err != nil {
			panic(fmt.Errorf("sql library returned non-base64 response"))
		}
		resp := &rpc.Response{}
		err = proto.Unmarshal(resMsg, resp)
		if err != nil {
			panic(fmt.Errorf("sql library returned non-proto response"))
		}

		// Done! Set response on job, and close its ch to notify Request
		j.resp = resp
		close(j.doneCh)
	}

	// We've exited the event loop, meaning Close() was called. Teardown the isolate and return.
	status, _, err = graalDetachAllThreadsAndTearDownIsolateFn.Call(uintptr(unsafe.Pointer(thread)))
	if status != 0 || err != nil {
		panic(fmt.Errorf("failed to teardown isolate"))
	}
}

// See getLibSQL
var (
	libsql     sharedlibrary.Library
	libsqlOnce sync.Once
)

// Returns a lazily-loaded reference to the SQL dynamic library
func getLibSQL() sharedlibrary.Library {
	// Lazily load libsql on first call
	libsqlOnce.Do(func() {
		// libraryFS and libraryPath are set in the platform-specific `deps_OS_ARCH` files in this package
		lib, err := sharedlibrary.OpenEmbed(libraryFS, libraryPath)
		if err != nil {
			panic(err)
		}
		libsql = lib
	})

	// If first call panic'ed, also panic on following calls (unlikely scenario)
	if libsql == nil {
		panic("libsql not loaded")
	}

	return libsql
}
