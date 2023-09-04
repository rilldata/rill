package schedule

import (
	"time"

	"github.com/rilldata/rill/runtime/pkg/priorityqueue"
)

// Schedule tracks unique values ordered by time.
// It is not thread-safe.
type Schedule[K comparable, V any] struct {
	hash  func(V) K
	pq    *priorityqueue.PriorityQueue[V]
	items map[K]*priorityqueue.Item[V]
}

// New creates a new Schedule. The hash function is used to determine a comparable key for a value.
func New[K comparable, V any](hash func(V) K) *Schedule[K, V] {
	return &Schedule[K, V]{
		hash:  hash,
		pq:    priorityqueue.New[V](true),
		items: make(map[K]*priorityqueue.Item[V]),
	}
}

// Set adds or updates the time of a value.
func (s Schedule[K, V]) Set(v V, t time.Time) {
	k := s.hash(v)
	i, ok := s.items[k]
	if ok {
		s.pq.Remove(i)
	}
	p := int(t.Unix())
	i = s.pq.Push(v, p)
	s.items[k] = i
}

// Remove removes a value from the schedule.
func (s Schedule[K, V]) Remove(v V) {
	k := s.hash(v)
	i, ok := s.items[k]
	if ok {
		s.pq.Remove(i)
		delete(s.items, k)
	}
}

// Pop removes the value with the earliest time from the schedule and returns it.
// It will panic if the schedule is empty.
func (s Schedule[K, V]) Pop() V {
	v := s.pq.Pop()
	delete(s.items, s.hash(v))
	return v
}

// Peek returns the value with the earliest time from the schedule.
func (s Schedule[K, V]) Peek() (V, time.Time) {
	if s.pq.Len() == 0 {
		var null V
		return null, time.Time{}
	}
	v := s.pq.Peek()
	i := s.items[s.hash(v)]
	return v, time.Unix(int64(i.Priority()), 0)
}

// Len returns the number of values in the schedule.
func (s Schedule[K, V]) Len() int {
	return len(s.items)
}
