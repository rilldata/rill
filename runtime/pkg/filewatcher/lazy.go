package filewatcher

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

type LazyWatcher struct {
	root        string
	ignorePaths []string
	logger      *zap.Logger

	mu      sync.Mutex
	count   int
	watcher *Watcher
}

func NewLazyWatcher(root string, ignorePaths []string, logger *zap.Logger) *LazyWatcher {
	return &LazyWatcher{
		root:        root,
		ignorePaths: ignorePaths,
		logger:      logger,
	}
}

func (w *LazyWatcher) Close() {
	w.mu.Lock()
	watcher := w.watcher
	w.watcher = nil
	w.mu.Unlock()
	if watcher != nil {
		watcher.Close()
	}
}

func (w *LazyWatcher) Subscribe(ctx context.Context, cb WatchCallback) error {
	w.mu.Lock()
	if w.watcher == nil {
		watcher, err := NewWatcher(w.root, w.ignorePaths, w.logger)
		if err != nil {
			w.mu.Unlock()
			return err
		}
		w.watcher = watcher
	}
	w.count++
	w.mu.Unlock()

	defer func() {
		w.mu.Lock()
		w.count--
		if w.count == 0 && w.watcher != nil {
			w.watcher.Close()
			w.watcher = nil
		}
		w.mu.Unlock()
	}()

	return w.watcher.Subscribe(ctx, cb)
}
