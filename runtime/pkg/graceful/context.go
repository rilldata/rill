package graceful

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// WithCancelOnTerminate derives a context that is cancelled on SIGINT and SIGTERM signals
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
