package logbuffer

import (
	"context"
	"sync"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func TestBufferGetLogs(t *testing.T) {
	b := NewBuffer(10, 4096)
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	require.NoError(t, b.AddEntry(runtimev1.LogLevel_LOG_LEVEL_DEBUG, t0, "debug", "{}"))
	require.NoError(t, b.AddEntry(runtimev1.LogLevel_LOG_LEVEL_INFO, t0.Add(time.Second), "info-1", "{}"))
	require.NoError(t, b.AddEntry(runtimev1.LogLevel_LOG_LEVEL_WARN, t0.Add(2*time.Second), "warn-1", "{}"))
	require.NoError(t, b.AddEntry(runtimev1.LogLevel_LOG_LEVEL_ERROR, t0.Add(3*time.Second), "error-1", "{}"))

	asc := b.GetLogs(true, -1, runtimev1.LogLevel_LOG_LEVEL_INFO)
	require.Len(t, asc, 3)
	require.Equal(t, []string{"info-1", "warn-1", "error-1"}, []string{asc[0].Message, asc[1].Message, asc[2].Message})

	desc := b.GetLogs(false, 2, runtimev1.LogLevel_LOG_LEVEL_INFO)
	require.Len(t, desc, 2)
	require.Equal(t, []string{"error-1", "warn-1"}, []string{desc[0].Message, desc[1].Message})
}

func TestBufferWatchLogs(t *testing.T) {
	b := NewBuffer(10, 4096)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	var got []string
	ready := make(chan struct{})
	done := make(chan error, 1)

	go func() {
		// Wait until WatchLogs has registered by polling or using a small handshake via AddEntry below.
		close(ready)
		done <- b.WatchLogs(ctx, func(item *runtimev1.Log) {
			mu.Lock()
			got = append(got, item.Message)
			mu.Unlock()
		}, runtimev1.LogLevel_LOG_LEVEL_WARN)
	}()

	<-ready
	require.Eventually(t, func() bool {
		b.mu.RLock()
		n := len(b.clients)
		b.mu.RUnlock()
		return n == 1
	}, 2*time.Second, 5*time.Millisecond)

	require.NoError(t, b.AddEntry(runtimev1.LogLevel_LOG_LEVEL_INFO, time.Now(), "ignored", "{}"))
	require.NoError(t, b.AddEntry(runtimev1.LogLevel_LOG_LEVEL_WARN, time.Now(), "warn-msg", "{}"))
	require.NoError(t, b.AddEntry(runtimev1.LogLevel_LOG_LEVEL_ERROR, time.Now(), "err-msg", "{}"))

	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(got) == 2
	}, 2*time.Second, 5*time.Millisecond)

	mu.Lock()
	require.Equal(t, []string{"warn-msg", "err-msg"}, got)
	mu.Unlock()

	cancel()
	err := <-done
	require.ErrorIs(t, err, context.Canceled)
}

// TestBufferWatchLogsCancelDoesNotDeadlockAddEntry verifies that cancelling a
// WatchLogs subscriber while AddEntry is concurrently broadcasting does not
// deadlock b.mu or block subsequent AddEntry/GetLogs calls.
func TestBufferWatchLogsCancelDoesNotDeadlockAddEntry(t *testing.T) {
	b := NewBuffer(100, 64*1024)

	ctx, cancel := context.WithCancel(context.Background())
	inCallback := make(chan struct{})
	releaseCallback := make(chan struct{})
	watchDone := make(chan error, 1)

	go func() {
		var once sync.Once
		watchDone <- b.WatchLogs(ctx, func(item *runtimev1.Log) {
			once.Do(func() {
				close(inCallback)
				<-releaseCallback
			})
		}, runtimev1.LogLevel_LOG_LEVEL_INFO)
	}()

	require.Eventually(t, func() bool {
		b.mu.RLock()
		n := len(b.clients)
		b.mu.RUnlock()
		return n == 1
	}, 2*time.Second, 5*time.Millisecond)

	// Trigger the first callback so WatchLogs is blocked inside fn.
	require.NoError(t, b.AddEntry(runtimev1.LogLevel_LOG_LEVEL_INFO, time.Now(), "first", "{}"))
	<-inCallback

	// Cancel WatchLogs while it is still inside fn, then send more entries.
	// With an unbuffered client channel held under b.mu, AddEntry would block
	// waiting for WatchLogs while WatchLogs blocks in removeClient waiting for b.mu.
	cancel()

	addDone := make(chan struct{})
	go func() {
		for i := 0; i < 256; i++ {
			_ = b.AddEntry(runtimev1.LogLevel_LOG_LEVEL_INFO, time.Now(), "burst", "{}")
		}
		close(addDone)
	}()

	select {
	case <-addDone:
	case <-time.After(2 * time.Second):
		close(releaseCallback)
		t.Fatal("AddEntry deadlocked while WatchLogs subscriber was blocked/cancelled")
	}

	close(releaseCallback)

	select {
	case err := <-watchDone:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(2 * time.Second):
		t.Fatal("WatchLogs did not exit after context cancellation")
	}

	// Buffer remains usable after subscriber removal.
	require.NotEmpty(t, b.GetLogs(true, 10, runtimev1.LogLevel_LOG_LEVEL_INFO))
}
