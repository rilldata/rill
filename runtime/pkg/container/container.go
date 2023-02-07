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

	// Full return true if container is full
	Full() bool

	// returns items as slice
	Items() []T
}

type bounded[T any] struct {
	items    []T
	count    int
	capacity int
}

// NewBounded returns a container of type bounded.
// Bounded is a container with limited capacity and
// stops consuming more elements once capacity is reached.
func NewBounded[T any](capacity int) (Container[T], error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("len should be greater than 0")
	}
	b := &bounded[T]{
		items:    make([]T, 0),
		count:    0,
		capacity: capacity,
	}
	return b, nil
}

func (b *bounded[T]) Add(item T) bool {
	if b.Full() {
		return false
	}

	b.items = append(b.items, item)
	b.count++
	return true
}

func (b *bounded[T]) Full() bool {
	return b.count == b.capacity
}

func (b *bounded[T]) Items() []T {
	return b.items[:b.count]
}

type fifo[T any] struct {
	items     *list.List
	capacity  int
	cleanupfn func(T)
}

// NewFIFO returns a container of type fifo.
// fifo is a container with limited capacity which keeps last 'N' inserted elements upto capacity.
// Additionally in order to do any cleanup on removed elements a cleanup function can be passed as well.
func NewFIFO[T any](capacity int, cleanupfn func(T)) (Container[T], error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("len should be greater than 0")
	}

	t := &fifo[T]{items: list.New(), capacity: capacity, cleanupfn: cleanupfn}
	return t, nil
}

func (t *fifo[T]) Add(item T) bool {
	if t.items.Len() == t.capacity {
		front := t.items.Remove(t.items.Back()).(T)
		if t.cleanupfn != nil {
			t.cleanupfn(front)
		}
	}
	t.items.PushFront(item)
	return true
}

func (t *fifo[T]) Full() bool {
	return false
}

func (t *fifo[T]) Items() []T {
	total := t.items.Len()
	result := make([]T, total)
	for i, curr := 0, t.items.Front(); curr != nil; i, curr = i+1, curr.Next() {
		result[i] = curr.Value.(T)
	}
	return result
}

type unbounded[T any] struct {
	items []T
}

// NewUnbounded returns a container of type unbounded.
// unbounded is a container with unlimited capacity.
func NewUnbounded[T any]() (Container[T], error) {
	t := &unbounded[T]{items: make([]T, 0)}
	return t, nil
}

func (u *unbounded[T]) Add(item T) bool {
	u.items = append(u.items, item)
	return true
}

func (u *unbounded[T]) Full() bool {
	// UnboundedContainer is never full
	return false
}

func (u *unbounded[T]) Items() []T {
	// create a copy ??
	return u.items
}
