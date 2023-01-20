package container

import (
	"container/list"
	"fmt"
)

// Container is a general purpose collection of items
// More functions like remove can be added if future use cases arise
// Can be made generic if more use cases come
type Container interface {
	// Add consumes an item
	Add(item any) bool

	// IsFull return true if container is full
	isFull() bool

	// returns items as slice
	Items() []any
}

// boundedContainer is a container with limited capacity
// Stops consuming more elements once capacity is reached
// This is thread unsafe
type boundedContainer struct {
	items []any
	index int
}

func NewBoundedContainer(capacity int) (Container, error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("len should be greater than 0")
	}
	b := &boundedContainer{items: make([]any, capacity), index: 0}
	return b, nil
}

func (b *boundedContainer) Add(item any) bool {
	if b.isFull() {
		return false
	}

	b.items[b.index] = item
	b.index++
	return true
}

func (b *boundedContainer) isFull() bool {
	return b.index == len(b.items)
}

func (b *boundedContainer) Items() []any {
	return b.items
}

// tailContainer is a container with limited capacity
// Keeps last 'N' inserted elements upto capacity
// This is thread unsafe
type tailContainer struct {
	items    *list.List
	capacity int
}

func NewTailContainer(capacity int) (Container, error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("len should be greater than 0")
	}

	t := &tailContainer{items: list.New(), capacity: capacity}
	return t, nil
}

func (b *tailContainer) Add(item any) bool {
	if b.isFull() {
		b.items.Remove(b.items.Back())
	}
	b.items.PushFront(item)
	return true
}

func (b *tailContainer) isFull() bool {
	return b.items.Len() == b.capacity
}

func (b *tailContainer) Items() []any {
	result := make([]any, b.items.Len())
	for i := 0; b.items.Len() > 0; i++ {
		front := b.items.Front()
		result[i] = front.Value
		b.items.Remove(front)
	}
	return result
}

// unboundedContainer is a container with unlimited capacity
// This is thread unsafe
type unboundedContainer struct {
	items []any
}

func NewUnboundedContainer() (Container, error) {
	t := &unboundedContainer{items: make([]any, 0)}
	return t, nil
}

func (b *unboundedContainer) Add(item any) bool {
	b.items = append(b.items, item)
	return true
}

func (b *unboundedContainer) isFull() bool {
	// UnboundedContainer is never full
	return false
}

func (b *unboundedContainer) Items() []any {
	// create a copy ??
	return b.items
}
