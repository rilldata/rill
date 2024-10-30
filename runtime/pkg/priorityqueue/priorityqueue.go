package priorityqueue

import (
	"container/heap"
)

// Item is a value in PriorityQueue
type Item[V any] struct {
	Value    V
	priority int
	index    int
}

func (i Item[V]) Priority() int {
	return i.priority
}

// PriorityQueue is a generic priority queue. It is not concurrency safe.
type PriorityQueue[V any] struct {
	heap priorityHeap[V]
}

// New creates a new priority queue.
// If minFirst is true, items with lower priority are returned first.
// If minFirst is false, items with higher priority are returned first.
func New[V any](minFirst bool) *PriorityQueue[V] {
	pq := &PriorityQueue[V]{}
	pq.heap.min = minFirst
	heap.Init(&pq.heap)
	return pq
}

func (pq *PriorityQueue[V]) Push(val V, priority int) *Item[V] {
	itm := &Item[V]{
		Value:    val,
		priority: priority,
		index:    -1,
	}
	heap.Push(&pq.heap, itm)
	return itm
}

func (pq *PriorityQueue[V]) Pop() V {
	itm := heap.Pop(&pq.heap).(*Item[V])
	return itm.Value
}

func (pq *PriorityQueue[V]) Peek() V {
	if len(pq.heap.items) > 0 {
		return pq.heap.items[0].Value
	}
	return Item[V]{}.Value // zero value
}

func (pq *PriorityQueue[V]) Remove(itm *Item[V]) {
	if itm.index >= 0 {
		heap.Remove(&pq.heap, itm.index)
	}
}

func (pq *PriorityQueue[V]) Contains(itm *Item[V]) bool {
	return itm.index >= 0
}

func (pq *PriorityQueue[V]) Len() int {
	return pq.heap.Len()
}

// priorityHeap implements heap.Interface to serve as a priority queue.
// See the heap docs for usage details: https://pkg.go.dev/container/heap#example-package-PriorityQueue
type priorityHeap[V any] struct {
	items []*Item[V]
	min   bool
}

func (pq priorityHeap[V]) Len() int { return len(pq.items) }

func (pq priorityHeap[V]) Less(i, j int) bool {
	if pq.min {
		return pq.items[i].priority < pq.items[j].priority
	}
	return pq.items[i].priority > pq.items[j].priority
}

func (pq priorityHeap[V]) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].index = i
	pq.items[j].index = j
}

func (pq *priorityHeap[V]) Push(x any) {
	n := len(pq.items)
	itm := x.(*Item[V])
	itm.index = n
	pq.items = append(pq.items, itm)
}

func (pq *priorityHeap[V]) Pop() any {
	old := pq.items
	n := len(old)
	itm := old[n-1]
	old[n-1] = nil // avoid memory leak
	itm.index = -1 // for safety
	pq.items = old[0 : n-1]
	return itm
}
