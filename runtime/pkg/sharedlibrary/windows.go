//go:build windows

package sharedlibrary

import (
	"syscall"
)

func Open(path string) (Library, error) {
	dll, err := syscall.LoadDLL(path)
	if err != nil {
		return nil, err
	}

	lib := &winLibrary{dll: dll}
	return lib, nil
}

type winLibrary struct {
	dll *syscall.DLL
}

func (l *winLibrary) FindFunc(name string) (Func, error) {
	proc, err := l.dll.FindProc(name)
	if err != nil {
		return nil, err
	}
	fn := &winFunc{proc: proc}
	return fn, nil
}

func (l *winLibrary) Close() error {
	return l.dll.Release()
}

type winFunc struct {
	proc *syscall.Proc
}

func (f *winFunc) Call(args ...uintptr) (uintptr, uintptr, error) {
	return f.proc.Call(args...)
}
