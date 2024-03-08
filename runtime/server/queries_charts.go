package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *Server) GetChartData(w http.ResponseWriter, req *http.Request) error {
	// TODO: telemetry

	ctx := req.Context()
	instanceId := req.PathValue("instance_id")
	chartName := req.PathValue("chart_name")
	// TODO: is a separate auth needed?
	if !auth.GetClaims(ctx).CanInstance(instanceId, auth.ReadMetrics) {
		return ErrForbidden
	}

	ch, err := resolveChart(ctx, s.runtime, instanceId, chartName)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return httputil.Errorf(http.StatusBadRequest, "failed to read request body: %w", err)
	}

	reqParams := make(map[string]interface{})
	if len(body) > 0 { // post
		if err := json.Unmarshal(body, &reqParams); err != nil {
			return httputil.Errorf(http.StatusBadRequest, "failed to unmarshal request body: %w", err)
		}
	}

	queryParams := req.URL.Query()
	for k, v := range queryParams {
		// set only the first value so that client does need to put array accessors in templates.
		reqParams[k] = v[0]
	}

	var api *runtimev1.API
	if ch.MetricsSql != "" {
		resolverPropsPB, err := structpb.NewStruct(map[string]interface{}{
			"sql": ch.MetricsSql,
		})
		if err != nil {
			httputil.Error(http.StatusInternalServerError, err)
		}
		// TODO: are other fields needed?
		api = &runtimev1.API{
			Spec: &runtimev1.APISpec{
				Resolver:           "Metrics",
				ResolverProperties: resolverPropsPB,
			},
		}
	} else if ch.Api != "" {
		api, err = s.runtime.APIForName(ctx, instanceId, ch.Api)
		if err != nil {
			if errors.Is(err, drivers.ErrResourceNotFound) {
				// shouldn't happen since validation would have happened in reconcile
				return httputil.Errorf(http.StatusNotFound, "api with name %q does not exist", ch.Api)
			}
			return httputil.Error(http.StatusInternalServerError, err)
		}
	}

	res, err := runtime.Resolve(ctx, &runtime.APIResolverOptions{
		Runtime:        s.runtime,
		InstanceID:     instanceId,
		API:            api,
		Args:           reqParams,
		UserAttributes: auth.GetClaims(ctx).Attributes(),
		Priority:       0,
	})

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(res)
	if err != nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}

	return nil
}

func resolveChart(ctx context.Context, rt *runtime.Runtime, instanceID, name string) (*runtimev1.ChartSpec, error) {
	ctrl, err := rt.Controller(ctx, instanceID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindChart, Name: name}, false)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ch := res.GetChart()
	spec := ch.Spec
	if spec == nil {
		return nil, status.Errorf(codes.InvalidArgument, "chart %q is invalid", name)
	}

	return spec, nil
}
