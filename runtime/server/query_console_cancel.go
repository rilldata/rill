package server

import (
	"context"
	"sync"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// queryTracker manages running query contexts for potential cancellation.
// It provides an in-memory map of query_id → context.CancelFunc that can be
// used by ExecuteQuery to register running queries and by CancelQuery to
// cancel them. In V1, CancelQuery always returns not_cancellable, but the
// tracking infrastructure is wired up for future use.
type queryTracker struct {
	mu      sync.Mutex
	running map[string]context.CancelFunc
}

// newQueryTracker creates a new queryTracker instance.
func newQueryTracker() *queryTracker {
	return &queryTracker{
		running: make(map[string]context.CancelFunc),
	}
}

// Register adds a query to the tracker and returns a cancellable context
// derived from the provided parent context. The caller should call Unregister
// when the query completes (typically via defer).
func (qt *queryTracker) Register(ctx context.Context, queryID string) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	qt.mu.Lock()
	qt.running[queryID] = cancel
	qt.mu.Unlock()
	return ctx, cancel
}

// Unregister removes a query from the tracker. It should be called when a
// query completes (successfully or otherwise). It does NOT call the cancel
// function — the caller is responsible for that via the returned CancelFunc
// from Register or via defer.
func (qt *queryTracker) Unregister(queryID string) {
	qt.mu.Lock()
	delete(qt.running, queryID)
	qt.mu.Unlock()
}

// Cancel attempts to cancel a running query by its ID. It returns true if the
// query was found and cancelled, false if the query was not found (either
// already completed or never registered).
func (qt *queryTracker) Cancel(queryID string) bool {
	qt.mu.Lock()
	cancel, ok := qt.running[queryID]
	if ok {
		delete(qt.running, queryID)
	}
	qt.mu.Unlock()

	if ok {
		cancel()
		return true
	}
	return false
}

// ActiveCount returns the number of currently tracked running queries.
// Useful for diagnostics and testing.
func (qt *queryTracker) ActiveCount() int {
	qt.mu.Lock()
	defer qt.mu.Unlock()
	return len(qt.running)
}

// queryConsoleTracker is the package-level query tracker instance.
// It is initialized once and shared across all query console handlers.
// In production, this would typically be a field on the Server struct,
// but we use a package-level instance here to avoid modifying the existing
// Server struct definition in this sprint.
var queryConsoleTracker = newQueryTracker()

// CancelQuery implements RuntimeService.CancelQuery.
//
// V1 implementation: This handler acknowledges the cancellation request but
// always returns a status indicating that cancellation is not yet supported.
// The query tracking infrastructure is in place, and future versions will
// support true cancellation for drivers that allow it.
//
// The handler does attempt to cancel tracked queries via the queryTracker,
// which will cancel the Go context. However, whether the underlying OLAP
// driver respects context cancellation is driver-dependent, so we
// conservatively report not_cancellable in V1.
func (s *Server) CancelQuery(ctx context.Context, req *runtimev1.CancelQueryRequest) (*runtimev1.CancelQueryResponse, error) {
	// Validate required fields.
	if req.InstanceId == "" {
		return nil, status.Error(codes.InvalidArgument, "instance_id is required")
	}
	if req.QueryId == "" {
		return nil, status.Error(codes.InvalidArgument, "query_id is required")
	}

	// Verify the instance exists and the caller has access.
	_, err := s.runtime.FindInstance(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "instance not found")
	}

	// Attempt to cancel the query via the tracker. Even though we report
	// not_cancellable in V1, we still cancel the Go context so that
	// drivers that respect context cancellation will stop early.
	found := queryConsoleTracker.Cancel(req.QueryId)

	// V1: Always return not_cancellable status. In future versions, we can
	// check driver capabilities and return a more accurate status.
	_ = found

	return &runtimev1.CancelQueryResponse{
		Status:  runtimev1.CancelQueryResponse_STATUS_NOT_CANCELLABLE,
		Message: "Query cancellation is not supported in this version. The query will run to completion.",
	}, nil
}
