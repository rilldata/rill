package debugserver

import (
	"context"
	"fmt"
	"net/http"

	// Register /debug/pprof/* endpoints on http.DefaultServeMux
	_ "net/http/pprof"
)

func ServeHTTP(ctx context.Context, port int) error {
	// A server without a handler will serve http.DefaultServeMux
	addr := ":" + fmt.Sprint(port)
	srv := &http.Server{Addr: addr}

	// Run server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	// Handle errors and context cancellation
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		_ = srv.Close()
		return ctx.Err()
	}
}
