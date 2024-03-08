package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *Server) GetChartData(w http.ResponseWriter, req *http.Request) error {
	// TODO: telemetry

	ctx := req.Context()
	instanceID := req.PathValue("instance_id")
	chartName := req.PathValue("chart_name")
	// TODO: is a separate auth needed?
	if !auth.GetClaims(ctx).CanInstance(instanceID, auth.ReadMetrics) {
		return httputil.Errorf(http.StatusForbidden, "does not have access to charts")
	}

	ch, err := resolveChart(ctx, s.runtime, instanceID, chartName)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return httputil.Errorf(http.StatusNotFound, "chart with name %s does not exist", chartName)
		}
		return httputil.Error(http.StatusInternalServerError, err)
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

	api, err := resolveAPI(ctx, s.runtime, instanceID, ch.Resolver, ch.ResolverProperties.AsMap())
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return httputil.Errorf(http.StatusNotFound, "api does not exist")
		}
		return httputil.Error(http.StatusInternalServerError, err)
	}

	res, err := runtime.Resolve(ctx, &runtime.APIResolverOptions{
		Runtime:        s.runtime,
		InstanceID:     instanceID,
		API:            api,
		Args:           reqParams,
		UserAttributes: auth.GetClaims(ctx).Attributes(),
		Priority:       0,
	})
	if err != nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}

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
		return nil, err
	}

	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindChart, Name: name}, false)
	if err != nil {
		return nil, err
	}

	ch := res.GetChart()
	spec := ch.Spec
	if spec == nil {
		return nil, fmt.Errorf("chart %q is invalid", name)
	}

	return spec, nil
}

func resolveAPI(ctx context.Context, rt *runtime.Runtime, instanceID, resolver string, resolverProps map[string]interface{}) (*runtimev1.API, error) {
	var api *runtimev1.API
	var err error

	switch resolver {
	case "Metrics":
		resolverPropsPB, err := structpb.NewStruct(map[string]interface{}{
			"sql": resolverProps["sql"],
		})
		if err != nil {
			return nil, err
		}
		// TODO: are other fields needed?
		api = &runtimev1.API{
			Spec: &runtimev1.APISpec{
				Resolver:           "Metrics",
				ResolverProperties: resolverPropsPB,
			},
		}

	case "API":
		apiName, ok := resolverProps["api"].(string)
		if !ok {
			return nil, errors.New("api name is missing")
		}
		api, err = rt.APIForName(ctx, instanceID, apiName)
		if err != nil {
			return nil, err
		}

	default:
		return nil, errors.New("unknown resolver")
	}

	return api, nil
}
