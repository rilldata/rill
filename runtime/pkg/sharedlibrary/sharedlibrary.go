package sharedlibrary

type Library interface {
	FindFunc(name string) (Func, error)
	Close() error
}

type Func interface {
	Call(args ...uintptr) (uintptr, uintptr, error)
}
