package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// SSEHandler TODO.
// This is required as vanguard doesn't currently map streaming RPCs to SSE, so we register this handler manually override the behavior
func (s *Server) SSEHandler(w http.ResponseWriter, req *http.Request) {
	// Parse the instance ID
	instanceID := req.PathValue("instance_id")

	// Parse the stream(s) to subscribe to.
	// The recommended way is to use the 'streams' parameter, which should be a comma-separated list.
	// This enables unified serving of multiple streams over one SSE connection.
	//
	// For backwards compatibility with event-specific endpoints, it also supports a 'stream' parameter with a single stream name.
	// When using 'stream', the response will omit the `event:` field in the SSE messages.
	var streams []string
	var omitEventNames bool
	q := req.URL.Query()
	if s := q.Get("streams"); s != "" {
		streams = strings.Split(s, ",")
	}
	if s := q.Get("stream"); s != "" {
		if len(streams) > 0 {
			http.Error(w, "cannot specify both 'stream' and 'streams' parameters", http.StatusBadRequest)
			return
		}
		streams = []string{s}
		omitEventNames = true
	}

	// Add observability attributes
	observability.AddRequestAttributes(req.Context(),
		attribute.String("args.instance_id", instanceID),
		attribute.StringSlice("args.streams", streams),
	)

	// Validation
	if len(streams) == 0 {
		http.Error(w, "must specify at least one stream via 'stream' or 'streams' parameter", http.StatusBadRequest)
		return
	}

	// Setup SSE handler
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()
	grp := errgroup.Group{} // Not using errgroup.WithContext since we want to send an error message before cancelling the SSE connection.
	sse := SSEHandler{
		Ctx:    ctx,
		Events: make(chan *SSEEvent),
	}

	// Start a goroutine with a streaming gRPC call for each requested stream.

	// Files stream
	if slices.Contains(streams, "files") {
		grp.Go(func() error {
			replayStr := req.URL.Query().Get("files_replay")
			replay, err := strconv.ParseBool(replayStr)
			if err != nil {
				return fmt.Errorf("invalid value for 'files_replay': %w", err)
			}

			rr := &runtimev1.WatchFilesRequest{
				InstanceId: instanceID,
				Replay:     replay,
			}

			ss := &grpcStreamingShim[*runtimev1.WatchFilesResponse]{
				ctx: ctx,
				fn: func(data []byte) error {
					event := &SSEEvent{Event: "files", Data: data}
					if omitEventNames {
						event.Event = ""
					}
					sse.Events <- event
					return nil
				},
			}

			return s.WatchFiles(rr, ss)
		})
	}

	// Resources stream
	if slices.Contains(streams, "resources") {
		grp.Go(func() error {
			kind := req.URL.Query().Get("resources_kind")

			replayStr := req.URL.Query().Get("resources_replay")
			replay, err := strconv.ParseBool(replayStr)
			if err != nil {
				return fmt.Errorf("invalid value for 'resources_replay': %w", err)
			}

			rr := &runtimev1.WatchResourcesRequest{
				InstanceId: instanceID,
				Kind:       kind,
				Replay:     replay,
			}

			ss := &grpcStreamingShim[*runtimev1.WatchResourcesResponse]{
				ctx: ctx,
				fn: func(data []byte) error {
					event := &SSEEvent{Event: "resources", Data: data}
					if omitEventNames {
						event.Event = ""
					}
					sse.Events <- event
					return nil
				},
			}

			return s.WatchResources(rr, ss)
		})
	}

	// Logs stream
	if slices.Contains(streams, "logs") {
		grp.Go(func() error {
			replayStr := req.URL.Query().Get("logs_replay")
			replay, err := strconv.ParseBool(replayStr)
			if err != nil {
				return fmt.Errorf("invalid value for 'logs_replay': %w", err)
			}

			var replayLimit int64
			if replayLimitStr := req.URL.Query().Get("logs_replay_limit"); replayLimitStr != "" {
				replayLimit, err = strconv.ParseInt(replayLimitStr, 10, 32)
				if err != nil {
					return fmt.Errorf("invalid value for 'logs_replay_limit': %w", err)
				}
			}

			var level runtimev1.LogLevel
			if levelStr := req.URL.Query().Get("logs_level"); levelStr != "" {
				level = runtimev1.LogLevel(runtimev1.LogLevel_value[strings.ToUpper(levelStr)])
			}

			rr := &runtimev1.WatchLogsRequest{
				InstanceId:  instanceID,
				Replay:      replay,
				ReplayLimit: int32(replayLimit),
				Level:       level,
			}

			ss := &grpcStreamingShim[*runtimev1.WatchLogsResponse]{
				ctx: ctx,
				fn: func(data []byte) error {
					event := &SSEEvent{Event: "logs", Data: data}
					if omitEventNames {
						event.Event = ""
					}
					sse.Events <- event
					return nil
				},
			}

			return s.WatchLogs(rr, ss)
		})
	}

	// Wait for all goroutines to complete
	go func() {
		err := grp.Wait()
		if err != nil && !errors.Is(err, context.Canceled) {
			s.logger.Warn("sse stream error", zap.String("instance_id", instanceID), zap.Error(err))
		}

		if err != nil {
			errJSON, err := json.Marshal(map[string]string{"error": err.Error()})
			if err != nil {
				s.logger.Error("failed to marshal error as json", zap.Error(err))
			}

			sse.Events <- &SSEEvent{
				Event: "error",
				Data:  errJSON,
			}
		}

		cancel()
		close(sse.Events)
	}()

	// Serve the SSE stream
	sse.ServeHTTP(w)
}

// SSEEvent represents a Server-Sent Event.
type SSEEvent struct {
	Event string
	Data  []byte
}

// ServeHTTP serves an SSE connection.
// It's implementation was adapted from github.com/r3labs/sse/v2.
type SSEHandler struct {
	Ctx    context.Context
	Events chan *SSEEvent
}

func (s *SSEHandler) ServeHTTP(w http.ResponseWriter) {
	// Check we support streaming responses.
	flusher, err := w.(http.Flusher)
	if !err {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// NOTE: Not supporting the Last-Event-ID header because our SSE connections are wrappers around underlying gRPC streams.
	// So any replay/cursor functionality should be implemented with dedicated parameters that feed into the underlying gRPC requests.

	// Send the headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// Push events
	for {
		select {
		case <-s.Ctx.Done():
			return
		case ev, ok := <-s.Events:
			// Exit if the events channel is closed
			if !ok {
				return
			}

			// Skip empty events
			if ev == nil || len(ev.Data) == 0 {
				continue
			}

			// Write the event
			if len(ev.Event) > 0 {
				_, err := fmt.Fprintf(w, "event: %s\n", ev.Event)
				if err != nil {
					return
				}
			}
			_, err := fmt.Fprintf(w, "data: %s\n", ev.Data)
			if err != nil {
				return
			}
			_, err = fmt.Fprint(w, "\n")
			if err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

// A shim for grpc.ServerStreamingServer that invokes a callback for response value serialized as JSON.
type grpcStreamingShim[Res proto.Message] struct {
	ctx context.Context
	fn  func(jsonData []byte) error
}

// Ensure grpcStreamingShim implements the grpc.ServerStreamingServer interface.
// (NOTE: The use of structpb.Value is not important, it's just a dummy value to satisfy the type parameter.)
var _ grpc.ServerStreamingServer[structpb.Value] = &grpcStreamingShim[*structpb.Value]{}

// Context returns the context of the request.
func (ss *grpcStreamingShim[Res]) Context() context.Context {
	return ss.ctx
}

// SendHeader sends a header to the client.
func (ss *grpcStreamingShim[Res]) Send(e Res) error {
	data, err := protojson.Marshal(e)
	if err != nil {
		return err
	}
	return ss.fn(data)
}

// SetHeader sets the header for the response.
func (ss *grpcStreamingShim[Res]) SetHeader(metadata.MD) error {
	return errors.New("not implemented")
}

// SendHeader sends a header to the client.
func (ss *grpcStreamingShim[Res]) SendHeader(metadata.MD) error {
	return errors.New("not implemented")
}

// SetTrailer sets the trailer for the response.
func (ss *grpcStreamingShim[Res]) SetTrailer(metadata.MD) {}

func (ss *grpcStreamingShim[Res]) SendMsg(m any) error {
	return errors.New("not implemented")
}

// RecvMsg receives a message from the client.
func (ss *grpcStreamingShim[Res]) RecvMsg(m any) error {
	return errors.New("not implemented")
}
