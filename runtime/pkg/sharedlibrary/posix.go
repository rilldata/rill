//go:build !windows && cgo

package sharedlibrary

/*
	#cgo LDFLAGS: -ldl

	#include <dlfcn.h>
	#include <limits.h>
	#include <stdlib.h>
	#include <stdint.h>
	#include <stdio.h>

	static uintptr_t lib_open(const char* path) {
		void* h = dlopen(path, RTLD_LAZY|RTLD_GLOBAL);
		if (h == NULL) {
			printf("dlopen err: %s\n", (char*)dlerror());
			return 0;
		}
		return (uintptr_t)h;
	}

	static uintptr_t lib_lookup(uintptr_t h, const char* name) {
		void* r = dlsym((void*)h, name);
		if (r == NULL) {
			printf("dlsym err: %s\n", (char*)dlerror());
			return 0;
		}
		return (uintptr_t)r;
	}

	static void lib_close(uintptr_t h) {
		if (h != 0) {
			dlclose((void*)h);
		}
	}

    static uint64_t call0(void* addr) {
		return ((uint64_t(*)())addr)();
	}

    static uint64_t call1(void* addr, void* p1) {
		return ((uint64_t(*)(void*))addr)(p1);
	}

    static uint64_t call2(void* addr, void* p1, void* p2) {
		return ((uint64_t(*)(void*,void*))addr)(p1, p2);
	}

    static uint64_t call3(void* addr, void* p1, void* p2, void* p3) {
		return ((uint64_t(*)(void*,void*,void*))addr)(p1, p2, p3);
	}

    static uint64_t call4(void* addr, void* p1, void* p2, void* p3, void* p4) {
		return ((uint64_t(*)(void*,void*,void*,void*))addr)(p1, p2, p3, p4);
	}

    static uint64_t call5(void* addr, void* p1, void* p2, void* p3, void* p4, void* p5) {
		return ((uint64_t(*)(void*,void*,void*,void*,void*))addr)(p1, p2, p3, p4, p5);
	}

    static uint64_t call6(void* addr, void* p1, void* p2, void* p3, void* p4, void* p5, void* p6) {
		return ((uint64_t(*)(void*,void*,void*,void*,void*,void*))addr)(p1, p2, p3, p4, p5, p6);
	}

    static uint64_t call7(void* addr, void* p1, void* p2, void* p3, void* p4, void* p5, void* p6, void *p7) {
		return ((uint64_t(*)(void*,void*,void*,void*,void*,void*, void*))addr)(p1, p2, p3, p4, p5, p6, p7);
	}

    static uint64_t call8(void* addr, void* p1, void* p2, void* p3, void* p4, void* p5, void* p6, void *p7, void *p8) {
		return ((uint64_t(*)(void*,void*,void*,void*,void*,void*,void*,void*))addr)(p1, p2, p3, p4, p5, p6, p7, p8);
	}

    static uint64_t call9(void* addr, void* p1, void* p2, void* p3, void* p4, void* p5, void* p6, void *p7, void *p8, void *p9) {
		return ((uint64_t(*)(void*,void*,void*,void*,void*,void*,void*,void*,void*))addr)(p1, p2, p3, p4, p5, p6, p7, p8, p9);
	}

    static uint64_t call10(void* addr, void* p1, void* p2, void* p3, void* p4, void* p5, void* p6, void *p7, void *p8, void *p9, void *p10) {
		return ((uint64_t(*)(void*,void*,void*,void*,void*,void*,void*,void*,void*,void*))addr)(p1, p2, p3, p4, p5, p6, p7, p8, p9, p10);
	}

    static uint64_t call11(void* addr, void* p1, void* p2, void* p3, void* p4, void* p5, void* p6, void *p7, void *p8, void *p9, void *p10, void *p11) {
		return ((uint64_t(*)(void*,void*,void*,void*,void*,void*,void*,void*,void*,void*,void*))addr)(p1, p2, p3, p4, p5, p6, p7, p8, p9, p10, p11);
	}

    static uint64_t call12(void* addr, void* p1, void* p2, void* p3, void* p4, void* p5, void* p6, void *p7, void *p8, void *p9, void *p10, void *p11, void *p12) {
		return ((uint64_t(*)(void*,void*,void*,void*,void*,void*,void*,void*,void*,void*,void*,void*))addr)(p1, p2, p3, p4, p5, p6, p7, p8, p9, p10, p11, p12);
	}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

func Open(path string) (Library, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	cRealPath := (*C.char)(C.malloc(C.PATH_MAX + 1))
	defer C.free(unsafe.Pointer(cRealPath))

	var handle C.uintptr_t
	if C.realpath(cPath, cRealPath) == nil {
		handle = C.lib_open(cPath)
	} else {
		handle = C.lib_open(cRealPath)
	}

	if handle == 0 {
		return nil, fmt.Errorf("can't open shared library '%s'", path)
	}

	lib := &posixLibrary{handle: handle}
	return lib, nil
}

type posixLibrary struct {
	handle C.uintptr_t
}

func (l *posixLibrary) FindFunc(name string) (Func, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	handle := C.lib_lookup(l.handle, cName)
	if handle == 0 {
		return nil, fmt.Errorf("can't find function '%s'", name)
	}

	fn := &posixFunc{handle: uintptr(handle)}
	return fn, nil
}

func (l *posixLibrary) Close() error {
	C.lib_close(l.handle)
	return nil
}

type posixFunc struct {
	handle uintptr
}

func (f *posixFunc) Call(args ...uintptr) (uintptr, uintptr, error) {
	var res C.uint64_t
	switch len(args) {
	case 0:
		res = C.call0(unsafe.Pointer(f.handle))
	case 1:
		res = C.call1(
			unsafe.Pointer(f.handle),
			unsafe.Pointer(args[0]),
		)
	case 2:
		res = C.call2(
			unsafe.Pointer(f.handle),
			unsafe.Pointer(args[0]),
			unsafe.Pointer(args[1]),
		)
	case 3:
		res = C.call3(
			unsafe.Pointer(f.handle),
			unsafe.Pointer(args[0]),
			unsafe.Pointer(args[1]),
			unsafe.Pointer(args[2]),
		)
	case 4:
		res = C.call4(
			unsafe.Pointer(f.handle),
			unsafe.Pointer(args[0]),
			unsafe.Pointer(args[1]),
			unsafe.Pointer(args[2]),
			unsafe.Pointer(args[3]),
		)
	case 5:
		res = C.call5(
			unsafe.Pointer(f.handle),
			unsafe.Pointer(args[0]),
			unsafe.Pointer(args[1]),
			unsafe.Pointer(args[2]),
			unsafe.Pointer(args[3]),
			unsafe.Pointer(args[4]),
		)
	case 6:
		res = C.call6(
			unsafe.Pointer(f.handle),
			unsafe.Pointer(args[0]),
			unsafe.Pointer(args[1]),
			unsafe.Pointer(args[2]),
			unsafe.Pointer(args[3]),
			unsafe.Pointer(args[4]),
			unsafe.Pointer(args[5]),
		)
	case 7:
		res = C.call7(
			unsafe.Pointer(f.handle),
			unsafe.Pointer(args[0]),
			unsafe.Pointer(args[1]),
			unsafe.Pointer(args[2]),
			unsafe.Pointer(args[3]),
			unsafe.Pointer(args[4]),
			unsafe.Pointer(args[5]),
			unsafe.Pointer(args[6]),
		)
	case 8:
		res = C.call8(
			unsafe.Pointer(f.handle),
			unsafe.Pointer(args[0]),
			unsafe.Pointer(args[1]),
			unsafe.Pointer(args[2]),
			unsafe.Pointer(args[3]),
			unsafe.Pointer(args[4]),
			unsafe.Pointer(args[5]),
			unsafe.Pointer(args[6]),
			unsafe.Pointer(args[7]),
		)
	case 9:
		res = C.call9(
			unsafe.Pointer(f.handle),
			unsafe.Pointer(args[0]),
			unsafe.Pointer(args[1]),
			unsafe.Pointer(args[2]),
			unsafe.Pointer(args[3]),
			unsafe.Pointer(args[4]),
			unsafe.Pointer(args[5]),
			unsafe.Pointer(args[6]),
			unsafe.Pointer(args[7]),
			unsafe.Pointer(args[8]),
		)
	case 10:
		res = C.call10(
			unsafe.Pointer(f.handle),
			unsafe.Pointer(args[0]),
			unsafe.Pointer(args[1]),
			unsafe.Pointer(args[2]),
			unsafe.Pointer(args[3]),
			unsafe.Pointer(args[4]),
			unsafe.Pointer(args[5]),
			unsafe.Pointer(args[6]),
			unsafe.Pointer(args[7]),
			unsafe.Pointer(args[8]),
			unsafe.Pointer(args[9]),
		)
	case 11:
		res = C.call11(
			unsafe.Pointer(f.handle),
			unsafe.Pointer(args[0]),
			unsafe.Pointer(args[1]),
			unsafe.Pointer(args[2]),
			unsafe.Pointer(args[3]),
			unsafe.Pointer(args[4]),
			unsafe.Pointer(args[5]),
			unsafe.Pointer(args[6]),
			unsafe.Pointer(args[7]),
			unsafe.Pointer(args[8]),
			unsafe.Pointer(args[9]),
			unsafe.Pointer(args[10]),
		)
	case 12:
		res = C.call12(
			unsafe.Pointer(f.handle),
			unsafe.Pointer(args[0]),
			unsafe.Pointer(args[1]),
			unsafe.Pointer(args[2]),
			unsafe.Pointer(args[3]),
			unsafe.Pointer(args[4]),
			unsafe.Pointer(args[5]),
			unsafe.Pointer(args[6]),
			unsafe.Pointer(args[7]),
			unsafe.Pointer(args[8]),
			unsafe.Pointer(args[9]),
			unsafe.Pointer(args[10]),
			unsafe.Pointer(args[11]),
		)
	default:
		panic(fmt.Errorf("sharedlibrary: cannot call function with more than 12 args"))
	}
	return uintptr(res), uintptr(res >> 32), nil
}
