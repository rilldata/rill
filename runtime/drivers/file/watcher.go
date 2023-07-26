package file

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

const batchInterval = 250 * time.Millisecond

const maxBufferSize = 1000

// watcher implements a recursive, batching file watcher on top of fsnotify.
type watcher struct {
	root        string
	watcher     *fsnotify.Watcher
	done        chan struct{}
	err         error
	mu          sync.Mutex
	subscribers map[string]drivers.WatchCallback
	buffer      []drivers.WatchEvent
}

func newWatcher(root string) (*watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &watcher{
		root:        root,
		watcher:     fsw,
		done:        make(chan struct{}),
		subscribers: make(map[string]drivers.WatchCallback),
	}

	err = w.addDir(root, false)
	if err != nil {
		w.watcher.Close()
		return nil, err
	}

	go w.run()

	return w, nil
}

func (w *watcher) close() {
	w.closeWithErr(nil)
}

func (w *watcher) closeWithErr(err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.done:
		// Already closed
		return
	default:
	}

	closeErr := w.watcher.Close()
	w.err = errors.Join(err, closeErr)
	if w.err == nil {
		w.err = fmt.Errorf("file watcher closed")
	}

	close(w.done)
}

func (w *watcher) subscribe(ctx context.Context, fn drivers.WatchCallback) error {
	w.mu.Lock()
	if w.err != nil {
		w.mu.Unlock()
		return w.err
	}
	id := fmt.Sprintf("%v", fn)
	w.subscribers[id] = fn
	w.mu.Unlock()

	defer func() {
		w.mu.Lock()
		delete(w.subscribers, id)
		w.mu.Unlock()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-w.done:
		return w.err
	}
}

// flush emits buffered events to all subscribers.
// Note it is called in the event loop in runInner, so new events will not be appended to w.buffer while a flush is running.
// Calls to flush block until all subscribers have processed the events. This is an acceptable trade-off for now, but we may want to revisit it in the future.
func (w *watcher) flush() {
	if len(w.buffer) == 0 {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	for _, fn := range w.subscribers {
		fn(w.buffer)
	}

	w.buffer = nil
}

func (w *watcher) run() {
	err := w.runInner()
	w.closeWithErr(err)
}

func (w *watcher) runInner() error {
	ticker := time.NewTicker(batchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ticker.Stop()
			w.flush()
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return nil
			}
			return err
		case e, ok := <-w.watcher.Events:
			if !ok {
				return nil
			}

			we := drivers.WatchEvent{}
			if e.Has(fsnotify.Create) || e.Has(fsnotify.Write) {
				we.Type = runtimev1.FileEvent_FILE_EVENT_WRITE
			} else if e.Has(fsnotify.Remove) || e.Has(fsnotify.Rename) {
				we.Type = runtimev1.FileEvent_FILE_EVENT_DELETE
			} else {
				continue
			}

			path, err := filepath.Rel(w.root, e.Name)
			if err != nil {
				return err
			}
			path = filepath.Join("/", path)
			we.Path = path

			if e.Has(fsnotify.Create) {
				info, err := os.Stat(e.Name)
				we.Dir = err == nil && info.IsDir()
			}

			w.buffer = append(w.buffer, we)

			// Calling addDir after appending to w.buffer, to sequence events correctly
			if we.Dir && e.Has(fsnotify.Create) {
				err = w.addDir(e.Name, true)
				if err != nil {
					return err
				}
			}

			// Reset the timer so we only flush when no events have been observed for batchInterval.
			// (But to avoid the buffer growing infinitely in edge cases, we enforce a max buffer size.)
			if len(w.buffer) < maxBufferSize {
				ticker.Reset(batchInterval)
			} else {
				ticker.Stop()
				w.flush()
			}
		}
	}
}

func (w *watcher) addDir(path string, replay bool) error {
	err := w.watcher.Add(path)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, e := range entries {
		fullPath := filepath.Join(path, e.Name())

		if replay {
			ep, err := filepath.Rel(w.root, fullPath)
			if err != nil {
				return err
			}
			ep = filepath.Join("/", ep)

			w.buffer = append(w.buffer, drivers.WatchEvent{
				Path: ep,
				Type: runtimev1.FileEvent_FILE_EVENT_WRITE,
				Dir:  e.IsDir(),
			})
		}

		if e.IsDir() {
			err := w.addDir(fullPath, replay)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
