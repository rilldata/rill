package ctxsync

import (
	"context"
	"math"

	"golang.org/x/sync/semaphore"
)

const semSize = math.MaxUint32

// RWMutex is similar to sync.RWMutex with handling for context cancellation in Lock and RLock.
type RWMutex struct {
	sem *semaphore.Weighted
}

func NewRWMutex() RWMutex {
	return RWMutex{sem: semaphore.NewWeighted(semSize)}
}

func (m RWMutex) Lock(ctx context.Context) error {
	return m.sem.Acquire(ctx, semSize)
}

func (m RWMutex) Unlock() {
	m.sem.Release(semSize)
}

func (m RWMutex) RLock(ctx context.Context) error {
	return m.sem.Acquire(ctx, 1)
}

func (m RWMutex) RUnlock() {
	m.sem.Release(1)
}
