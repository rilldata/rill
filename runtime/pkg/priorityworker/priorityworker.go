package priorityworker

import (
	"container/heap"
	"context"
	"errors"
	"sync"
)

// ErrStopped is returned for outstanding items when the queue is stopped
var ErrStopped = errors.New("priorityqueue: stopped")

// Handler is a callback called by PriorityWorker to process an item
type Handler[V any] func(context.Context, V) error

// PriorityWorker implements a concurrency-safe worker that prioritizes work
// using a priority queue.
type PriorityWorker[V any] struct {
	handler      Handler[V]
	enqueueJobCh chan *item[V]
	cancelJobCh  chan *item[V]
	stopped      bool
	mu           sync.RWMutex
	stopDoneCh   chan struct{}
}

// New creates a new PriorityWorker that calls the provided handler for every
// item submitted using Process. It starts a goroutine for the worker's event
// loop. You must call Stop on the worker when you're done using it.
func New[V any](handler Handler[V]) *PriorityWorker[V] {
	pw := &PriorityWorker[V]{
		handler:      handler,
		enqueueJobCh: make(chan *item[V]),
		cancelJobCh:  make(chan *item[V]),
		stopped:      false,
		mu:           sync.RWMutex{},
		stopDoneCh:   make(chan struct{}),
	}

	go pw.work()

	return pw
}

// Process enqueues a work item and returns when the item has been dequeued and processed.
// It returns early if ctx is cancelled or the worker is stopped.
func (pw *PriorityWorker[V]) Process(ctx context.Context, priority int, val V) error {
	job := &item[V]{
		ctx:      ctx,
		doneCh:   make(chan struct{}),
		val:      val,
		priority: priority,
		index:    -1,
	}

	pw.mu.RLock()
	if !pw.stopped {
		pw.enqueueJobCh <- job
	} else {
		job.err = ErrStopped
		close(job.doneCh)
	}
	pw.mu.RUnlock()

	select {
	case <-job.doneCh:
	case <-ctx.Done():
		pw.cancelJobCh <- job
		job.err = context.Canceled
	}

	return job.err
}

func (pw *PriorityWorker[V]) Stop() {
	pw.mu.Lock()
	pw.stopped = true
	pw.mu.Unlock()
	close(pw.enqueueJobCh)
	<-pw.stopDoneCh
	return
}

func (pw *PriorityWorker[V]) work() {
	pq := priorityQueue[V]{}
	heap.Init(&pq)

	var currentDoneCh chan struct{}

	for {
		select {
		case job, ok := <-pw.enqueueJobCh:
			// If enqueueJobCh was stopped, it means it's time to stop
			if !ok {
				// Cancel all enqueued items
				for _, job := range pq {
					job.err = ErrStopped
					close(job.doneCh)
				}
				// Let the current item finish
				if currentDoneCh != nil {
					<-currentDoneCh
				}
				// Exit
				close(pw.stopDoneCh)
				return
			}

			// Process or enqueue item
			if currentDoneCh == nil {
				// If we're currently idle, we process the item directly
				currentDoneCh = job.doneCh
				go pw.handle(job)
			} else {
				// Else we add it to the priority queue
				heap.Push(&pq, job)
			}
		case job := <-pw.cancelJobCh:
			// Remove item from heap if it hasn't already been dequeued
			if job.index >= 0 {
				heap.Remove(&pq, job.index)
			}
		case <-currentDoneCh:
			if pq.Len() > 0 {
				// If the queue isn't empty, start a new item
				job := heap.Pop(&pq).(*item[V])
				currentDoneCh = job.doneCh
				go pw.handle(job)
			} else {
				// Else, go idle until next query
				currentDoneCh = nil
			}
		}
	}
}

func (pw *PriorityWorker[V]) handle(job *item[V]) {
	// Bail if the job's ctx is cancelled
	// (Unlikely to happen given other safeguards)
	if job.ctx.Err() != nil {
		close(job.doneCh)
		return
	}

	// Run
	err := pw.handler(job.ctx, job.val)
	job.err = err
	close(job.doneCh)
}

// item represents a job enqueued in priorityQueue
type item[V any] struct {
	// Worker related fields
	ctx    context.Context // ctx from caller
	doneCh chan struct{}   // used to indicate the job has finished
	val    V               // value the job should work on
	err    error           // err to return for the job

	// Priority queue related fields
	priority int
	index    int
}

// priorityQueue implements heap.Interface to serve as a priority queue of jobs
// See the docs for details: https://pkg.go.dev/container/heap#example-package-PriorityQueue
type priorityQueue[V any] []*item[V]

func (pq priorityQueue[V]) Len() int { return len(pq) }

func (pq priorityQueue[V]) Less(i, j int) bool {
	// We use greater than here so that Pop gives us the highest priority item (not lowest)
	return pq[i].priority > pq[j].priority
}

func (pq priorityQueue[V]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue[V]) Push(x any) {
	n := len(*pq)
	itm := x.(*item[V])
	itm.index = n
	*pq = append(*pq, itm)
}

func (pq *priorityQueue[V]) Pop() any {
	old := *pq
	n := len(old)
	itm := old[n-1]
	old[n-1] = nil // avoid memory leak
	itm.index = -1 // for safety
	*pq = old[0 : n-1]
	return itm
}
