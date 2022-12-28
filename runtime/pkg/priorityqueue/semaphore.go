package priorityqueue

import (
	"context"
	"sync"
)

// Semaphore implements a counting semaphore that's acquired in prioritized order.
// The implementation is derived from golang.org/x/sync/semaphore.
type Semaphore struct {
	mu   sync.Mutex
	pq   *PriorityQueue[chan struct{}]
	size int
	cur  int
}

// NewSemaphore creates a Semaphore where size is the maximum.
func NewSemaphore(size int) *Semaphore {
	return &Semaphore{
		mu:   sync.Mutex{},
		pq:   New[chan struct{}](),
		size: size,
		cur:  0,
	}
}

// Acquire acquires the semaphore with a priority. Higher priorities are acquired first.
// It blocks until the semaphore is acquired or ctx is cancelled.
// If ctx is cancelled, Acquire returns ctx.Err(), otherwise it always returns nil.
func (s *Semaphore) Acquire(ctx context.Context, priority int) error {
	s.mu.Lock()
	if s.size-s.cur >= 1 && s.pq.Len() == 0 {
		s.cur++
		s.mu.Unlock()
		return nil
	}

	readyCh := make(chan struct{})
	itm := s.pq.Push(readyCh, priority)
	s.mu.Unlock()

	select {
	case <-readyCh:
		return nil
	case <-ctx.Done():
		s.mu.Lock()
		if !s.pq.Contains(itm) {
			// Cancelled and acquired at the same time. Easiest to pretend it was acquired first.
			s.mu.Unlock()
			return nil
		}
		s.pq.Remove(itm)
		s.mu.Unlock()
		return ctx.Err()
	}
}

// TryAcquire tries to immediately acquire the semaphore.
// It returns false if the semaphore is locked or there are items in the queue.
func (s *Semaphore) TryAcquire() bool {
	s.mu.Lock()
	ok := s.size-s.cur >= 1 && s.pq.Len() == 0
	if ok {
		s.cur++
	}
	s.mu.Unlock()
	return ok
}

// Release releases a semaphore previously acquired with Acquire or TryAcquire.
func (s *Semaphore) Release() {
	s.mu.Lock()
	s.cur--
	if s.cur < 0 {
		s.mu.Unlock()
		panic("semaphore released more times than acquired")
	}
	s.notifyWaiters()
	s.mu.Unlock()
}

// notifyWaiters pops items off the priority queue until the semaphore is full or the queue is empty.
// It must be called while `s.mu` is locked.
func (s *Semaphore) notifyWaiters() {
	for {
		if s.pq.Len() == 0 {
			break
		}

		if s.cur == s.size {
			break
		}

		readyCh := s.pq.Pop()
		s.cur++
		close(readyCh)
	}
}
