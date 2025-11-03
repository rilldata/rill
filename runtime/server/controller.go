package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/r3labs/sse/v2"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/pagination"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListResources implements runtimev1.RuntimeServiceServer
func (s *Server) ListResources(ctx context.Context, req *runtimev1.ListResourcesRequest) (*runtimev1.ListResourcesResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.kind", req.Kind),
		attribute.Bool("args.skip_security_checks", req.SkipSecurityChecks),
	)

	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadObjects) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	rs, err := ctrl.List(ctx, req.Kind, req.Path, false)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	slices.SortFunc(rs, func(a, b *runtimev1.Resource) int {
		an := a.Meta.Name
		bn := b.Meta.Name
		if an.Kind < bn.Kind {
			return -1
		}
		if an.Kind > bn.Kind {
			return 1
		}
		return strings.Compare(an.Name, bn.Name)
	})

	if req.SkipSecurityChecks {
		if !claims.Admin() {
			return nil, ErrForbidden
		}
		return &runtimev1.ListResourcesResponse{Resources: rs}, nil
	}

	i := 0
	for i < len(rs) {
		r := rs[i]
		r, access, err := s.runtime.ApplySecurityPolicy(ctx, req.InstanceId, claims, r)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if !access {
			// Remove from the slice
			rs[i] = rs[len(rs)-1]
			rs[len(rs)-1] = nil
			rs = rs[:len(rs)-1]
			continue
		}
		rs[i] = r
		i++
	}

	return &runtimev1.ListResourcesResponse{Resources: rs}, nil
}

// WatchResources implements runtimev1.RuntimeServiceServer
func (s *Server) WatchResources(req *runtimev1.WatchResourcesRequest, ss runtimev1.RuntimeService_WatchResourcesServer) error {
	observability.AddRequestAttributes(ss.Context(),
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.kind", req.Kind),
	)

	claims := auth.GetClaims(ss.Context(), req.InstanceId)
	if !claims.Can(runtime.ReadObjects) {
		return ErrForbidden
	}

	ctrl, err := s.runtime.Controller(ss.Context(), req.InstanceId)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if req.Replay {
		rs, err := ctrl.List(ss.Context(), req.Kind, "", false)
		if err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}

		for _, r := range rs {
			r, access, err := s.runtime.ApplySecurityPolicy(ss.Context(), req.InstanceId, claims, r)
			if err != nil {
				return status.Error(codes.InvalidArgument, err.Error())
			}
			if !access {
				continue
			}

			err = ss.Send(&runtimev1.WatchResourcesResponse{
				Event:    runtimev1.ResourceEvent_RESOURCE_EVENT_WRITE,
				Resource: r,
			})
			if err != nil {
				return status.Error(codes.InvalidArgument, err.Error())
			}
		}
	}

	return ctrl.Subscribe(ss.Context(), func(e runtimev1.ResourceEvent, n *runtimev1.ResourceName, r *runtimev1.Resource) {
		if r != nil { // r is nil for deletion events
			var access bool
			var err error
			r, access, err = s.runtime.ApplySecurityPolicy(ss.Context(), req.InstanceId, claims, r)
			if err != nil {
				s.logger.Info("failed to apply security policy", zap.String("name", n.Name), zap.Error(err))
				return
			}
			if !access {
				return
			}
		}

		err = ss.Send(&runtimev1.WatchResourcesResponse{
			Event:    e,
			Name:     n,
			Resource: r,
		})
		if err != nil {
			s.logger.Info("failed to send resource event", zap.Error(err))
		}
	})
}

// GetResource implements runtimev1.RuntimeServiceServer
func (s *Server) GetResource(ctx context.Context, req *runtimev1.GetResourceRequest) (*runtimev1.GetResourceResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.name.kind", req.Name.Kind),
		attribute.String("args.name.name", req.Name.Name),
		attribute.Bool("args.skip_security_checks", req.SkipSecurityChecks),
	)

	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadObjects) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	r, err := ctrl.Get(ctx, req.Name, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.SkipSecurityChecks {
		if !claims.Admin() {
			return nil, ErrForbidden
		}
		return &runtimev1.GetResourceResponse{Resource: r}, nil
	}

	r, access, err := s.runtime.ApplySecurityPolicy(ctx, req.InstanceId, claims, r)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if !access {
		return nil, status.Error(codes.NotFound, "resource not found")
	}

	return &runtimev1.GetResourceResponse{Resource: r}, nil
}

// GetExplore implements runtimev1.RuntimeServiceServer
func (s *Server) GetExplore(ctx context.Context, req *runtimev1.GetExploreRequest) (*runtimev1.GetExploreResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.name", req.Name),
	)

	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadObjects) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	n := &runtimev1.ResourceName{Kind: runtime.ResourceKindExplore, Name: req.Name}
	e, err := ctrl.Get(ctx, n, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	e, access, err := s.runtime.ApplySecurityPolicy(ctx, req.InstanceId, claims, e)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if !access {
		return nil, status.Error(codes.NotFound, "resource not found")
	}

	validSpec := e.GetExplore().State.ValidSpec
	if validSpec == nil {
		return &runtimev1.GetExploreResponse{
			Explore: e,
		}, nil
	}

	n = &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: validSpec.MetricsView}
	m, err := ctrl.Get(ctx, n, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, status.Error(codes.NotFound, "metrics view not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	m, access, err = s.runtime.ApplySecurityPolicy(ctx, req.InstanceId, claims, m)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if !access {
		return nil, status.Error(codes.NotFound, "metrics view not found")
	}

	return &runtimev1.GetExploreResponse{
		Explore:     e,
		MetricsView: m,
	}, nil
}

// GetModelPartitions implements runtimev1.RuntimeServiceServer
func (s *Server) GetModelPartitions(ctx context.Context, req *runtimev1.GetModelPartitionsRequest) (*runtimev1.GetModelPartitionsResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.model", req.Model),
	)

	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadObjects) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	n := &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: req.Model}
	r, err := ctrl.Get(ctx, n, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	r, access, err := s.runtime.ApplySecurityPolicy(ctx, req.InstanceId, claims, r)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if !access {
		return nil, status.Error(codes.NotFound, "resource not found")
	}

	partitionsModelID := r.GetModel().State.PartitionsModelId
	if partitionsModelID == "" {
		return &runtimev1.GetModelPartitionsResponse{}, nil
	}

	var beforeExecutedOn time.Time
	afterKey := ""
	if req.PageToken != "" {
		err := pagination.UnmarshalPageToken(req.PageToken, &beforeExecutedOn, &afterKey)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to parse page token: %v", err)
		}
	}

	catalog, release, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	defer release()

	defaultPageSize := 100
	opts := &drivers.FindModelPartitionsOptions{
		ModelID:          partitionsModelID,
		WherePending:     req.Pending,
		WhereErrored:     req.Errored,
		BeforeExecutedOn: beforeExecutedOn,
		AfterKey:         afterKey,
		Limit:            pagination.ValidPageSize(req.PageSize, defaultPageSize),
	}

	partitions, err := catalog.FindModelPartitions(ctx, opts)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var nextPageToken string
	if len(partitions) == pagination.ValidPageSize(req.PageSize, defaultPageSize) {
		last := partitions[len(partitions)-1]
		nextPageToken = pagination.MarshalPageToken(last.Index, last.Key)
	}

	return &runtimev1.GetModelPartitionsResponse{
		Partitions:    modelPartitionsToPB(partitions),
		NextPageToken: nextPageToken,
	}, nil
}

// CreateTrigger implements runtimev1.RuntimeServiceServer
func (s *Server) CreateTrigger(ctx context.Context, req *runtimev1.CreateTriggerRequest) (*runtimev1.CreateTriggerResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.EditTrigger) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Build refresh trigger spec
	spec := &runtimev1.RefreshTriggerSpec{
		Resources: req.Resources,
		Models:    req.Models,
	}

	// Handle the convenience flag for the project parser.
	if req.Parser {
		spec.Resources = append(spec.Resources, runtime.GlobalProjectParserName)
	}

	// Handle the convenience flags for all (user-facing) resources.
	// In practice, we only refresh the major user declared resources that impact serving.
	// For example, we don't currently trigger alerts or reports.
	if req.All || req.AllFull {
		kinds := []string{
			runtime.ResourceKindProjectParser,
			runtime.ResourceKindConnector,
			runtime.ResourceKindModel,
			runtime.ResourceKindMetricsView,
			runtime.ResourceKindExplore,
			runtime.ResourceKindComponent,
			runtime.ResourceKindCanvas,
		}
		for _, kind := range kinds {
			rs, err := ctrl.List(ctx, kind, "", false)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, fmt.Errorf("failed to list resources of kind %q: %w", kind, err).Error())
			}
			for _, r := range rs {
				if kind == runtime.ResourceKindModel {
					spec.Models = append(spec.Models, &runtimev1.RefreshModelTrigger{
						Model: r.Meta.Name.Name,
						Full:  req.AllFull,
					})
					continue
				}
				spec.Resources = append(spec.Resources, r.Meta.Name)
			}
		}
	}

	// Create the trigger resource
	name := fmt.Sprintf("trigger_%s", randomString(8))
	n := &runtimev1.ResourceName{Kind: runtime.ResourceKindRefreshTrigger, Name: name}
	r := &runtimev1.Resource{Resource: &runtimev1.Resource_RefreshTrigger{RefreshTrigger: &runtimev1.RefreshTrigger{Spec: spec}}}
	err = ctrl.Create(ctx, n, nil, nil, nil, false, r)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Errorf("failed to create trigger: %w", err).Error())
	}

	return &runtimev1.CreateTriggerResponse{}, nil
}

// WatchResourcesHandler implements an HTTP handler for runtimev1.RuntimeServiceServer
func (s *Server) WatchResourcesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	instanceID := r.PathValue("instance_id")
	kind := r.URL.Query().Get("kind")
	replay := r.URL.Query().Get("replay") == "true"

	claims := auth.GetClaims(ctx, instanceID)
	if !claims.Can(runtime.ReadObjects) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	eventServer := sse.New()
	eventServer.CreateStream("resources")
	eventServer.Headers = map[string]string{
		"Content-Type":  "text/event-stream",
		"Cache-Control": "no-cache",
		"Connection":    "keep-alive",
	}

	// Create a shim that adapts the SSE server to the WatchResources gRPC server
	shim := &watchResourcesServerShim{r: r, sse: eventServer}

	// Use the existing WatchResources implementation in a goroutine
	go func() {
		err := s.WatchResources(&runtimev1.WatchResourcesRequest{
			InstanceId: instanceID,
			Kind:       kind,
			Replay:     replay,
		}, shim)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				s.logger.Warn("watch resources error", zap.String("instance_id", instanceID), zap.String("kind", kind), zap.Error(err))
			}

			errJSON, err := json.Marshal(map[string]string{"error": err.Error()})
			if err != nil {
				s.logger.Error("failed to marshal error as json", zap.Error(err))
			}

			eventServer.Publish("resources", &sse.Event{
				Data:  errJSON,
				Event: []byte("error"),
			})
		}
		eventServer.Close()
	}()

	eventServer.ServeHTTP(w, r)
}

// modelPartitionsToPB converts a slice of drivers.ModelPartition to a slice of runtimev1.ModelPartition.
func modelPartitionsToPB(partitions []drivers.ModelPartition) []*runtimev1.ModelPartition {
	pbs := make([]*runtimev1.ModelPartition, len(partitions))
	for i, partition := range partitions {
		pbs[i] = modelPartitionToPB(partition)
	}
	return pbs
}

// modelPartitionToPB converts a drivers.ModelPartition to a runtimev1.ModelPartition.
func modelPartitionToPB(partition drivers.ModelPartition) *runtimev1.ModelPartition {
	var data map[string]interface{}
	if err := json.Unmarshal(partition.DataJSON, &data); err != nil {
		panic(err)
	}

	var watermark, executedOn *timestamppb.Timestamp
	if partition.Watermark != nil {
		watermark = timestamppb.New(*partition.Watermark)
	}
	if partition.ExecutedOn != nil {
		executedOn = timestamppb.New(*partition.ExecutedOn)
	}

	return &runtimev1.ModelPartition{
		Key:        partition.Key,
		Data:       must(structpb.NewStruct(data)),
		Watermark:  watermark,
		ExecutedOn: executedOn,
		Error:      partition.Error,
		ElapsedMs:  uint32(partition.Elapsed.Milliseconds()),
	}
}

func randomString(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// A shim for runtimev1.RuntimeService_WatchResourcesServer
type watchResourcesServerShim struct {
	r   *http.Request
	sse *sse.Server
}

// Context returns the context from the HTTP request
func (s *watchResourcesServerShim) Context() context.Context {
	return s.r.Context()
}

// Send adapts the WatchResourcesResponse to SSE events
func (s *watchResourcesServerShim) Send(e *runtimev1.WatchResourcesResponse) error {
	data, err := protojson.Marshal(e)
	if err != nil {
		return err
	}

	s.sse.Publish("resources", &sse.Event{Data: data})
	return nil
}

// SetHeader implements the grpc.ServerStream interface
func (s *watchResourcesServerShim) SetHeader(metadata.MD) error {
	return nil // No-op for HTTP/SSE
}

// SendHeader implements the grpc.ServerStream interface
func (s *watchResourcesServerShim) SendHeader(metadata.MD) error {
	return nil // No-op for HTTP/SSE
}

// SetTrailer implements the grpc.ServerStream interface
func (s *watchResourcesServerShim) SetTrailer(metadata.MD) {
	// No-op for HTTP/SSE
}

// SendMsg implements the grpc.ServerStream interface
func (s *watchResourcesServerShim) SendMsg(m any) error {
	return errors.New("not implemented")
}

// RecvMsg implements the grpc.ServerStream interface
func (s *watchResourcesServerShim) RecvMsg(m any) error {
	return errors.New("not implemented")
}
