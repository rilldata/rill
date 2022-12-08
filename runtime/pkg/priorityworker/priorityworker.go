package priorityworker

import (
	"container/heap"
	"context"
	"errors"
	"golang.org/x/sync/semaphore"
	"sync"
)

// ErrStopped is returned for outstanding items when the queue is stopped.
var ErrStopped = errors.New("priorityqueue: stopped")

// Handler is a callback called by PriorityWorker to process an item.
type Handler[V any] func(context.Context, V) error

// PriorityWorker implements a concurrency-safe worker that prioritizes work
// using a priority queue.
type PriorityWorker[V any] struct {
	handler        Handler[V]
	enqueueJobCh   chan *item[V]
	cancelJobCh    chan *item[V]
	paused         bool
	pausedToggleCh chan bool
	stopped        bool
	stoppedMu      sync.RWMutex
	stopDoneCh     chan struct{}
	concurrency    int
	runningJobs    map[*item[V]]bool
}

// New creates a new PriorityWorker that calls the provided handler for every
// item submitted using Process. It starts a goroutine for the worker's event
// loop. You must call Stop on the worker when you're done using it.
func New[V any](handler Handler[V], concurrency int) *PriorityWorker[V] {
	pw := &PriorityWorker[V]{
		handler:        handler,
		enqueueJobCh:   make(chan *item[V]),
		cancelJobCh:    make(chan *item[V]),
		paused:         false,
		pausedToggleCh: make(chan bool),
		stopped:        false,
		stoppedMu:      sync.RWMutex{},
		stopDoneCh:     make(chan struct{}),
		concurrency:    concurrency,
		runningJobs:    make(map[*item[V]]bool),
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

	pw.stoppedMu.RLock()
	if !pw.stopped {
		pw.enqueueJobCh <- job
	} else {
		job.err = ErrStopped
		close(job.doneCh)
	}
	pw.stoppedMu.RUnlock()

	select {
	case <-job.doneCh:
	case <-ctx.Done():
		pw.cancelJobCh <- job
		job.err = context.Canceled
	}

	return job.err
}

// Pause keeps the queue open for new jobs, but won't process them until Unpause is called.
// Pause is mainly useful for predictable testing of the priority worker.
func (pw *PriorityWorker[V]) Pause() {
	pw.pausedToggleCh <- true
}

// Unpause reverses Pause.
func (pw *PriorityWorker[V]) Unpause() {
	pw.pausedToggleCh <- false
}

// Stop cancels all jobs that haven't started, and returns once the current job has finished.
func (pw *PriorityWorker[V]) Stop() {
	pw.stoppedMu.Lock()
	pw.stopped = true
	pw.stoppedMu.Unlock()
	close(pw.enqueueJobCh)
	<-pw.stopDoneCh
}

func (pw *PriorityWorker[V]) work() {
	pq := priorityQueue[V]{}
	heap.Init(&pq)

	jobDoneCh := make(chan struct{})
	sem := semaphore.NewWeighted(int64(pw.concurrency))

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
				// Wait for all running jobs to finish
				for job := range pw.runningJobs {
					<-job.doneCh
				}

				// Exit
				close(pw.stopDoneCh)
				return
			}

			// Process or enqueue item
			// check count of running jobs here and if less than concurrency then start otherwise push in queue
			if !pw.paused && sem.TryAcquire(1) == true {
				// If we're currently idle, we process the item directly
				go pw.handle(job, jobDoneCh, sem)
			} else {
				// Else we add it to the priority queue
				heap.Push(&pq, job)
			}
		case job := <-pw.cancelJobCh:
			// Remove item from heap if it hasn't already been dequeued
			if job.index >= 0 {
				heap.Remove(&pq, job.index)
			}
		case <-jobDoneCh:
			// some job completed, so we are sure that we have a free slot since channel will get messages in order,
			// but we need to acquire semaphore to make sure that we don't start more than concurrency jobs
			if !pw.paused && pq.Len() > 0 && sem.TryAcquire(1) == true {
				// If the queue isn't empty, start a new item
				job := heap.Pop(&pq).(*item[V])
				//currentDoneCh = job.doneCh
				go pw.handle(job, jobDoneCh, sem)
			}
		case p := <-pw.pausedToggleCh:
			pw.paused = p
			if !pw.paused && pq.Len() > 0 && sem.TryAcquire(1) == true {
				// We just unpaused, we're idle, and the queue is not empty – start the next job
				job := heap.Pop(&pq).(*item[V])
				go pw.handle(job, jobDoneCh, sem)
			}
		}
	}
}

func (pw *PriorityWorker[V]) cleanUpJobState(job *item[V], jobDoneCh chan struct{}, sem *semaphore.Weighted) {
	delete(pw.runningJobs, job)
	sem.Release(1)
	close(job.doneCh)
	jobDoneCh <- struct{}{}
}

func (pw *PriorityWorker[V]) handle(job *item[V], jobDoneCh chan struct{}, sem *semaphore.Weighted) {
	pw.runningJobs[job] = true
	defer pw.cleanUpJobState(job, jobDoneCh, sem)
	// Bail if the job's ctx is cancelled
	// (Unlikely to happen given other safeguards)
	if job.ctx.Err() != nil {
		job.err = job.ctx.Err()
		close(job.doneCh)
		return
	}

	// Run
	job.err = pw.handler(job.ctx, job.val)
}

// item represents a job enqueued in priorityQueue.
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
