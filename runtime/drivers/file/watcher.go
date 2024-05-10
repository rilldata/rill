package file

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"golang.org/x/sys/unix"
)

const batchInterval = 250 * time.Millisecond

const maxBufferSize = 1000

// watcher implements a recursive, batching file watcher on top of fsnotify.
type watcher struct {
	logger           *zap.Logger
	root             string
	ignorePaths      []string
	watcher          *fsnotify.Watcher
	closed           atomic.Bool
	done             chan struct{}
	err              error
	mu               sync.Mutex
	subscribers      map[int]drivers.WatchCallback
	nextSubscriberID int
	buffer           map[string]watchEvent
}

type watchEvent struct {
	eventType runtimev1.FileEvent
	path      string
	relPath   string
	dir       bool
	isCreate  bool
}

// newWatcher creates a new watcher for the given root directory.
// The root directory must be an absolute path.
func newWatcher(root string, ignorePaths []string, logger *zap.Logger) (*watcher, error) {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &watcher{
		logger:      logger,
		root:        root,
		ignorePaths: ignorePaths,
		watcher:     fsw,
		done:        make(chan struct{}),
		subscribers: make(map[int]drivers.WatchCallback),
		buffer:      make(map[string]watchEvent),
	}

	err = w.addDir(root, false, true)
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
	// Support multiple calls, but only actually close once.
	// Not using w.mu here because someday someone will try to close the watcher from a callback.
	if w.closed.Swap(true) {
		return
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
	id := w.nextSubscriberID
	w.nextSubscriberID++
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

	for p, event := range w.buffer {
		if !event.isCreate {
			continue
		}
		// check for directory for CREATE events
		info, err := os.Stat(event.path)
		event.dir = err == nil && info.IsDir()
		if event.dir {
			// add directory to tracking paths
			err = w.addDir(event.path, true, false)
			if err != nil {
				delete(w.buffer, p)
				continue
			}
		}
		w.buffer[p] = event
	}

	events := maps.Values(w.buffer)
	driverEvents := make([]drivers.WatchEvent, len(events))
	for i, event := range events {
		driverEvents[i] = drivers.WatchEvent{
			Type: event.eventType,
			Path: event.relPath,
			Dir:  event.dir,
		}
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	for _, fn := range w.subscribers {
		fn(driverEvents)
	}

	w.buffer = make(map[string]watchEvent)
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
			if err == nil || isNotExists(err) {
				continue
			}
			return err
		case e, ok := <-w.watcher.Events:
			if !ok {
				return nil
			}

			we := watchEvent{}
			if e.Has(fsnotify.Remove) || e.Has(fsnotify.Rename) {
				we.eventType = runtimev1.FileEvent_FILE_EVENT_DELETE
			} else if e.Has(fsnotify.Create) || e.Has(fsnotify.Write) || e.Has(fsnotify.Chmod) {
				we.eventType = runtimev1.FileEvent_FILE_EVENT_WRITE
			} else {
				continue
			}
			we.isCreate = e.Has(fsnotify.Create)

			path, err := filepath.Rel(w.root, e.Name)
			if err != nil {
				w.logger.Warn("ignoring watcher event: failed to get relative path", zap.String("root", w.root), zap.String("event_name", e.Name), zap.String("event_op", e.Op.String()))
				continue
			}

			path = filepath.Join("/", path)
			we.relPath = path
			we.path = e.Name

			// Do not send files for ignored paths
			if drivers.IsIgnored(path, w.ignorePaths) {
				continue
			}

			existing, ok := w.buffer[path]
			if ok && existing.isCreate && we.eventType == runtimev1.FileEvent_FILE_EVENT_WRITE {
				// copy over `IsCreate` within the batch for a path
				we.isCreate = existing.isCreate
			}
			w.buffer[path] = we

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

func (w *watcher) addDir(path string, replay, errIfNotExist bool) error {
	err := w.watcher.Add(path)
	if err != nil {
		// Need to check unix.ENOENT (and probably others) since fsnotify doesn't always use cross-platform syscalls.
		if !errIfNotExist && isNotExists(err) {
			return nil
		}
		return err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		if !errIfNotExist && isNotExists(err) {
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

			w.buffer[ep] = watchEvent{
				path:      fullPath,
				relPath:   ep,
				eventType: runtimev1.FileEvent_FILE_EVENT_WRITE,
				dir:       e.IsDir(),
			}
		}

		if e.IsDir() {
			err := w.addDir(fullPath, replay, errIfNotExist)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isNotExists(err error) bool {
	return os.IsNotExist(err) || errors.Is(err, unix.ENOENT)
}
