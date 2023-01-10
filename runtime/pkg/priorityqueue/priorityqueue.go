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

// PriorityQueue is a generic priority queue.
// It returns items with a higher priority first. It is not concurrency safe.
type PriorityQueue[V any] struct {
	heap priorityHeap[V]
}

func New[V any]() *PriorityQueue[V] {
	pq := &PriorityQueue[V]{}
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
type priorityHeap[V any] []*Item[V]

func (pq priorityHeap[V]) Len() int { return len(pq) }

func (pq priorityHeap[V]) Less(i, j int) bool {
	// We use greater than here so that Pop gives us the highest priority item (not lowest)
	return pq[i].priority > pq[j].priority
}

func (pq priorityHeap[V]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityHeap[V]) Push(x any) {
	n := len(*pq)
	itm := x.(*Item[V])
	itm.index = n
	*pq = append(*pq, itm)
}

func (pq *priorityHeap[V]) Pop() any {
	old := *pq
	n := len(old)
	itm := old[n-1]
	old[n-1] = nil // avoid memory leak
	itm.index = -1 // for safety
	*pq = old[0 : n-1]
	return itm
}
