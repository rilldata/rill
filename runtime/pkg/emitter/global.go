package emitter

import (
	"sync/atomic"
	"unsafe"
)

var globalClient unsafe.Pointer

func Get() *Client {
	return (*Client)(atomic.LoadPointer(&globalClient))
}

func Set(client *Client) {
	atomic.StorePointer(&globalClient, unsafe.Pointer(client))
}
