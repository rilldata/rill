package usage

import (
	"sync/atomic"
	"unsafe"
)

var globalClient unsafe.Pointer

func GetClient() *Client {
	return (*Client)(atomic.LoadPointer(&globalClient))
}

func SetClient(client *Client) {
	atomic.StorePointer(&globalClient, unsafe.Pointer(client))
}
