package runtime

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrTooManyConcurrentQueries is returned when a query could not get a slot to run within the allowed wait time.
var ErrTooManyConcurrentQueries = errors.New("too many concurrent queries")

// AcquireHeavyQuerySlot waits for a slot to run a heavy query on the instance.
// What counts as a heavy query is decided by the caller; for example, the metrics executor counts queries that page deep into a large result.
// At most instanceLimit heavy queries run concurrently on the instance, and at most userLimit of them for the same user.
// A limit of 0 or less disables that limit, and an empty userID skips the per-user limit.
// If no slot frees up within maxWait, it returns an error wrapping ErrTooManyConcurrentQueries.
// The caller must call the returned release function once the query's results have been consumed; extra calls are no-ops.
func (r *Runtime) AcquireHeavyQuerySlot(ctx context.Context, instanceID, userID string, instanceLimit, userLimit int, maxWait time.Duration) (func(), error) {
	return r.heavyQueries.acquire(ctx, instanceID, userID, instanceLimit, userLimit, maxWait)
}

// queryLimiter caps the number of queries that run concurrently per instance and per user within an instance.
// It keeps a single caller, such as an agent paginating through a large result in parallel, from using up an instance's memory and OLAP capacity.
// Waiting callers are not served in FIFO order; every release wakes all waiters, and the first to re-check gets the slot.
type queryLimiter struct {
	mu        sync.Mutex
	instances map[string]*instanceQuerySlots
}

// instanceQuerySlots tracks the slots in use on one instance.
type instanceQuerySlots struct {
	total  int
	byUser map[string]int
	// released is closed and replaced on every release, which wakes up the waiters so they can check again.
	released chan struct{}
}

func newQueryLimiter() *queryLimiter {
	return &queryLimiter{instances: make(map[string]*instanceQuerySlots)}
}

func (l *queryLimiter) acquire(ctx context.Context, instanceID, userID string, instanceLimit, userLimit int, maxWait time.Duration) (func(), error) {
	timer := time.NewTimer(maxWait)
	defer timer.Stop()

	for {
		l.mu.Lock()
		s, ok := l.instances[instanceID]
		if !ok {
			s = &instanceQuerySlots{
				byUser:   make(map[string]int),
				released: make(chan struct{}),
			}
			l.instances[instanceID] = s
		}

		instanceFull := instanceLimit > 0 && s.total >= instanceLimit
		userFull := userID != "" && userLimit > 0 && s.byUser[userID] >= userLimit
		if !instanceFull && !userFull {
			s.total++
			if userID != "" {
				s.byUser[userID]++
			}
			l.mu.Unlock()

			var once sync.Once
			return func() { once.Do(func() { l.release(instanceID, userID) }) }, nil
		}
		released := s.released
		l.mu.Unlock()

		select {
		case <-released:
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
			if userFull {
				return nil, fmt.Errorf("%w: this user already has %d heavy queries running (limit %d per user)", ErrTooManyConcurrentQueries, userLimit, userLimit)
			}
			return nil, fmt.Errorf("%w: the project already has %d heavy queries running (limit %d per project)", ErrTooManyConcurrentQueries, instanceLimit, instanceLimit)
		}
	}
}

func (l *queryLimiter) release(instanceID, userID string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	s := l.instances[instanceID]
	s.total--
	if userID != "" {
		s.byUser[userID]--
		if s.byUser[userID] == 0 {
			delete(s.byUser, userID)
		}
	}

	close(s.released)
	if s.total == 0 {
		delete(l.instances, instanceID)
	} else {
		s.released = make(chan struct{})
	}
}
