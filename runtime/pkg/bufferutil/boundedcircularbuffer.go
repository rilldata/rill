package bufferutil

import (
	"errors"
)

type BoundedCircularBuffer[T any] struct {
	data        []Item[T]
	capacity    int
	maxSize     int64
	count       int
	currentSize int64
	head        int
	tail        int
	zero        Item[T]
}

type Item[T any] struct {
	Value T
	Size  int
}

// NewBoundedCircularBuffer creates a new bounded circular buffer with the given capacity and max size. capacity is the
// number of items that can be stored in the buffer. maxSize is the total size of all items that can be stored in the
// buffer. If the buffer is full, the oldest items will be dropped until there is enough space for the new item.
// This buffer is not concurrency safe.
func NewBoundedCircularBuffer[T any](capacity int, maxSize int64) *BoundedCircularBuffer[T] {
	return &BoundedCircularBuffer[T]{
		data:        make([]Item[T], capacity),
		capacity:    capacity,
		maxSize:     maxSize,
		count:       0,
		currentSize: 0,
		head:        0,
		tail:        0,
		zero:        Item[T]{},
	}
}

func (cb *BoundedCircularBuffer[T]) Push(item Item[T]) {
	// Drop items from the head until there's enough space for the new item
	for cb.count > 0 && (cb.count == cb.capacity || int64(item.Size)+cb.currentSize > cb.maxSize) {
		cb.currentSize -= int64(cb.data[cb.tail].Size)
		cb.data[cb.tail] = cb.zero
		cb.tail = (cb.tail + 1) % cb.capacity
		cb.count--
	}

	cb.data[cb.head] = item
	cb.currentSize += int64(item.Size)
	cb.head = (cb.head + 1) % cb.capacity
	cb.count++
}

func (cb *BoundedCircularBuffer[T]) Pop() (Item[T], error) {
	if cb.count == 0 {
		return cb.zero, errors.New("buffer is empty")
	}
	item := cb.data[cb.tail]
	cb.data[cb.tail] = cb.zero
	cb.currentSize -= int64(item.Size)
	cb.tail = (cb.tail + 1) % cb.capacity
	cb.count--
	return item, nil
}

func (cb *BoundedCircularBuffer[T]) Peek() (Item[T], error) {
	if cb.count == 0 {
		return cb.zero, errors.New("buffer is empty")
	}
	item := cb.data[cb.tail]
	return item, nil
}

// Iterate iterates over the buffer from oldest to newest. The callback function will be called for each item in the
// buffer. The limit parameter specifies the maximum number of items to iterate over. No validation on limit is done.
// Caller is responsible for ensuring limit is a valid value.
func (cb *BoundedCircularBuffer[T]) Iterate(callback func(item Item[T]), limit int) {
	itemsToSkip := cb.count - limit
	pos := (cb.tail + itemsToSkip) % cb.capacity
	for i := 0; i < limit; i++ {
		callback(cb.data[pos])
		pos = (pos + 1) % cb.capacity
	}
}

// ReverseIterate iterates over the buffer from newest to oldest. The callback function will be called for each item in
// the buffer. The limit parameter specifies the maximum number of items to iterate over. No validation on limit is done.
// Caller is responsible for ensuring limit is a valid value.
func (cb *BoundedCircularBuffer[T]) ReverseIterate(callback func(item Item[T]), limit int) {
	pos := cb.head
	for i := 0; i < limit; i++ {
		pos = (pos - 1) % cb.capacity
		if pos < 0 {
			pos += cb.capacity
		}
		callback(cb.data[pos])
	}
}

// ReverseIterateUntil iterates over the buffer from oldest to newest until the callback function returns false, or it has
// iterated over the entire buffer.
func (cb *BoundedCircularBuffer[T]) ReverseIterateUntil(callback func(item Item[T]) bool) {
	pos := cb.head
	for i := 0; i < cb.count; i++ {
		pos = (pos - 1) % cb.capacity
		if pos < 0 {
			pos += cb.capacity
		}
		if !callback(cb.data[pos]) {
			return
		}
	}
}

func (cb *BoundedCircularBuffer[T]) Count() int {
	return cb.count
}
