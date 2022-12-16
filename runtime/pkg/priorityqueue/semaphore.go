package priorityqueue

import (
	"context"
	"sync"
)

type Semaphore struct {
	mu   sync.Mutex
	pq   *PriorityQueue[chan struct{}]
	size int
	cur  int
}

func NewSemaphore(size int) *Semaphore {
	return &Semaphore{
		mu:   sync.Mutex{},
		pq:   New[chan struct{}](),
		size: size,
		cur:  0,
	}
}

func (s *Semaphore) Acquire(ctx context.Context, priority int) error {
	s.mu.Lock()
	if s.size-s.cur >= 1 && s.pq.Len() == 0 {
		s.cur += 1
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

func (s *Semaphore) TryAcquire() bool {
	s.mu.Lock()
	ok := s.size-s.cur >= 1 && s.pq.Len() == 0
	if ok {
		s.cur += 1
	}
	s.mu.Unlock()
	return ok
}

func (s *Semaphore) Release() {
	s.mu.Lock()
	s.cur -= 1
	if s.cur < 0 {
		s.mu.Unlock()
		panic("semaphore released more times than acquired")
	}
	s.notifyWaiters()
	s.mu.Unlock()
}

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
