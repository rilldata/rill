package graceful

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// WithCancelOnTerminate derives a context that is cancelled on SIGINT and SIGTERM signals.
func WithCancelOnTerminate(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	// handles SIGINT and SIGTERM gracefully
	go func() {
		defer cancel()
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-c:
		case <-ctx.Done():
		}
	}()

	return ctx
}

// WithMinimumDuration derives a context that delays the parent's cancellation until the provided minimum duration has elapsed.
// When done with the derived context, call the returned cancel function to clean up associated resources.
func WithMinimumDuration(parentCtx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	newCtx, cancel := context.WithCancel(context.Background())

	go func() {
		// Wait until the minimum duration has elapsed.
		select {
		case <-newCtx.Done():
			return
		case <-time.After(d):
		}

		// Wait until the parent context is done.
		select {
		case <-newCtx.Done():
			return
		case <-parentCtx.Done():
			cancel()
		}
	}()

	return newCtx, cancel
}
