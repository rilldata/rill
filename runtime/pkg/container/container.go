package container

import (
	"container/list"
	"fmt"
)

// Container is a collection of items
// More functions like remove can be added if future use cases arise
type Container[T any] interface {
	// Add consumes an item
	Add(item T) bool

	// IsFull return true if container is full
	IsFull() bool

	// returns items as slice
	Items() []T
}

// boundedContainer is a container with limited capacity
// Stops consuming more elements once capacity is reached
// This is thread unsafe
type boundedContainer[T any] struct {
	items []T
	index int
}

func NewBoundedContainer[T any](capacity int) (Container[T], error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("len should be greater than 0")
	}
	b := &boundedContainer[T]{items: make([]T, capacity), index: 0}
	return b, nil
}

func (b *boundedContainer[T]) Add(item T) bool {
	if b.IsFull() {
		return false
	}

	b.items[b.index] = item
	b.index++
	return true
}

func (b *boundedContainer[T]) IsFull() bool {
	return b.index == len(b.items)
}

func (b *boundedContainer[T]) Items() []T {
	return b.items[:b.index]
}

// tailContainer is a container with limited capacity
// Keeps last 'N' inserted elements upto capacity
// This is thread unsafe
type tailContainer[T any] struct {
	items     *list.List
	capacity  int
	cleanupfn func(T)
}

func NewTailContainer[T any](capacity int, cleanupfn func(T)) (Container[T], error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("len should be greater than 0")
	}

	t := &tailContainer[T]{items: list.New(), capacity: capacity, cleanupfn: cleanupfn}
	return t, nil
}

func (t *tailContainer[T]) Add(item T) bool {
	if t.items.Len() == t.capacity {
		t.cleanupfn(t.items.Remove(t.items.Back()).(T))
	}
	t.items.PushFront(item)
	return true
}

func (t *tailContainer[T]) IsFull() bool {
	return false
}

func (t *tailContainer[T]) Items() []T {
	total := t.items.Len()
	result := make([]T, total)
	for i, curr := 0, t.items.Front(); curr != nil; i, curr = i+1, curr.Next() {
		result[i] = curr.Value.(T)
	}
	return result
}

// unboundedContainer is a container with unlimited capacity
// This is thread unsafe
type unboundedContainer[T any] struct {
	items []T
}

func NewUnboundedContainer[T any]() (Container[T], error) {
	t := &unboundedContainer[T]{items: make([]T, 0)}
	return t, nil
}

func (u *unboundedContainer[T]) Add(item T) bool {
	u.items = append(u.items, item)
	return true
}

func (u *unboundedContainer[T]) IsFull() bool {
	// UnboundedContainer is never full
	return false
}

func (u *unboundedContainer[T]) Items() []T {
	// create a copy ??
	return u.items
}
