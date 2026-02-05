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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	// sseEventFile is the SSE event type for events from WatchFiles.
	sseEventFile = "file"
	// sseEventResource is the SSE event type for events from WatchResources.
	sseEventResource = "resource"
	// sseEventLog is the SSE event type for events from WatchLogs.
	sseEventLog = "log"
)

// SSEHandler is a shim that exposes a unified SSE endpoint for the Server's streaming gRPC RPCs.
// This is useful for two reasons:
// 1. The vanguard library doesn't currently map streaming RPCs to SSE, so we need to manually implement that.
// 2. We need to provide a unified endpoint that can multiplex multiple streams over a single SSE connection. This helps support multiple open tabs from a single client since browsers limit concurrent unsecured/localhost connections to 6 per origin.
//
// The caller can differentiate between the event types using the `event:` field in the SSE messages.
//
// The handler supports the following query parameters:
// - events: comma-separated list of events to subscribe to. Supported values are: file (maps to WatchFiles), resource (maps to WatchResources), log (maps to WatchLogs).
// - stream: deprecated, use `events` instead. If used, the `event:` message field is omitted in the response for backwards compatibility. Supported values are: files, resources, logs.
// - files_replay: maps to WatchFilesRequest.Replay
// - resources_kind: maps to WatchResourcesRequest.Kind
// - resources_replay: maps to WatchResourcesRequest.Replay
// - logs_replay: maps to WatchLogsRequest.Replay
// - logs_replay_limit: maps to WatchLogsRequest.ReplayLimit
// - logs_level: maps to WatchLogsRequest.Level
func (s *Server) SSEHandler(w http.ResponseWriter, req *http.Request) {
	// Parse the instance ID
	instanceID := req.PathValue("instance_id")

	// Parse the event(s) to subscribe to.
	var eventTypes []string
	var omitEventNames bool
	q := req.URL.Query()
	if v := q.Get("events"); v != "" {
		eventTypes = strings.Split(v, ",")
	}
	if v := q.Get("stream"); v != "" { // For backwards compatibility, see function comment.
		if len(eventTypes) > 0 {
			http.Error(w, "cannot specify both 'stream' and 'events' parameters", http.StatusBadRequest)
			return
		}
		omitEventNames = true
		switch v {
		case "files":
			eventTypes = []string{sseEventFile}
		case "resources":
			eventTypes = []string{sseEventResource}
		case "logs":
			eventTypes = []string{sseEventLog}
		default:
			http.Error(w, fmt.Sprintf("unknown stream type %q", v), http.StatusBadRequest)
			return
		}
	}

	// Deduplicate to prevent starting multiple goroutines for the same event type.
	slices.Sort(eventTypes)
	eventTypes = slices.Compact(eventTypes)

	// Add observability attributes
	observability.AddRequestAttributes(req.Context(),
		attribute.String("args.instance_id", instanceID),
		attribute.StringSlice("args.events", eventTypes),
	)

	// Validation
	if len(eventTypes) == 0 {
		http.Error(w, "must specify at least one event type via the 'events' parameter", http.StatusBadRequest)
		return
	}

	// Start goroutines for each event type.
	grp, ctx := errgroup.WithContext(req.Context())
	events := make(chan *sseEvent)
	for _, eventType := range eventTypes {
		switch eventType {
		case sseEventFile:
			grp.Go(func() error {
				var replay bool
				if replayStr := req.URL.Query().Get("files_replay"); replayStr != "" {
					var err error
					replay, err = strconv.ParseBool(replayStr)
					if err != nil {
						return fmt.Errorf("invalid value for 'files_replay': %w", err)
					}
				}

				rr := &runtimev1.WatchFilesRequest{
					InstanceId: instanceID,
					Replay:     replay,
				}

				ss := &grpcStreamingShim[*runtimev1.WatchFilesResponse]{
					ctx: ctx,
					fn: func(data []byte) error {
						event := &sseEvent{Event: sseEventFile, Data: data}
						if omitEventNames {
							event.Event = ""
						}
						events <- event
						return nil
					},
				}

				return s.WatchFiles(rr, ss)
			})

		case sseEventResource:
			grp.Go(func() error {
				kind := req.URL.Query().Get("resources_kind")

				var replay bool
				if replayStr := req.URL.Query().Get("resources_replay"); replayStr != "" {
					var err error
					replay, err = strconv.ParseBool(replayStr)
					if err != nil {
						return fmt.Errorf("invalid value for 'resources_replay': %w", err)
					}
				}

				rr := &runtimev1.WatchResourcesRequest{
					InstanceId: instanceID,
					Kind:       kind,
					Replay:     replay,
				}

				ss := &grpcStreamingShim[*runtimev1.WatchResourcesResponse]{
					ctx: ctx,
					fn: func(data []byte) error {
						event := &sseEvent{Event: sseEventResource, Data: data}
						if omitEventNames {
							event.Event = ""
						}
						events <- event
						return nil
					},
				}

				return s.WatchResources(rr, ss)
			})

		case sseEventLog:
			grp.Go(func() error {
				var replay bool
				if replayStr := req.URL.Query().Get("logs_replay"); replayStr != "" {
					var err error
					replay, err = strconv.ParseBool(replayStr)
					if err != nil {
						return fmt.Errorf("invalid value for 'logs_replay': %w", err)
					}
				}

				var replayLimit int64
				if replayLimitStr := req.URL.Query().Get("logs_replay_limit"); replayLimitStr != "" {
					var err error
					replayLimit, err = strconv.ParseInt(replayLimitStr, 10, 32)
					if err != nil {
						return fmt.Errorf("invalid value for 'logs_replay_limit': %w", err)
					}
				}

				var level runtimev1.LogLevel
				if levelStr := req.URL.Query().Get("logs_level"); levelStr != "" {
					level = runtimev1.LogLevel(runtimev1.LogLevel_value[levelStr])
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
						event := &sseEvent{Event: sseEventLog, Data: data}
						if omitEventNames {
							event.Event = ""
						}
						events <- event
						return nil
					},
				}

				return s.WatchLogs(rr, ss)
			})

		default:
			http.Error(w, fmt.Sprintf("unknown event type: %s", eventType), http.StatusBadRequest)
			return
		}
	}

	// In the background, wait for goroutines to complete and send a final error event if applicable.
	// Attention must be paid to ctx handling and cancellation here. At a high-level, there are two scenarios:
	// 1. The request is cancelled. The ctx used by the streams is cancelled, so they return with context.Canceled. The grp.Wait() returns, this goroutine closes the events channel, making serveSSEUntilClose return.
	// 2. An error occurs in a stream. The errgroup cancels the ctx, so the other streams also returns. The grp.Wait() returns the original error, which this goroutine sends as a final message, then closes the events channel, making serveSSEUntilClose return.
	go func() {
		// This goroutine must close the events channel to ensure the call to serveSSEUntilClose returns.
		defer close(events)

		err := grp.Wait()
		if err != nil && !errors.Is(err, context.Canceled) {
			s.logger.Warn("sse stream error", zap.String("instance_id", instanceID), zap.Error(err))
		}

		if err != nil {
			code := codes.Unknown
			msg := err.Error()
			if s, ok := status.FromError(err); ok {
				code = s.Code()
				msg = s.Message()
			}

			errJSON, err := json.Marshal(map[string]string{"code": code.String(), "error": msg})
			if err != nil {
				s.logger.Error("failed to marshal error as json", zap.Error(err))
			}

			events <- &sseEvent{
				Event: "error",
				Data:  errJSON,
			}
		}
	}()

	// Serve the SSE stream.
	// This will only return when the background goroutine calls close(events).
	serveSSEUntilClose(w, events)
}

// sseEvent represents a Server-Sent Event.
type sseEvent struct {
	Event string
	Data  []byte
}

// serveSSEUntilClose serves SSE events from the provided channel until it's closed.
// Its implementation was adapted from github.com/r3labs/sse/v2.
func serveSSEUntilClose(w http.ResponseWriter, events chan *sseEvent) {
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

	// Consume events from channel and write to response (the loop ends when the channel is closed)
	for ev := range events {
		// Skip empty events
		if ev == nil || len(ev.Data) == 0 {
			continue
		}

		// Write the event
		if ev.Event != "" {
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
