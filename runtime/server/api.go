package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
)

// nolint
func (s *Server) APIForName(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
	if !auth.GetClaims(req.Context()).CanInstance(pathParams["instance_id"], auth.ReadOLAP) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ctx := req.Context()
	if pathParams["name"] == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	reqParams := make(map[string]interface{})
	if len(body) > 0 { // post
		if err := json.Unmarshal(body, &reqParams); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	queryParams := req.URL.Query()
	for k, v := range queryParams {
		// set only the first value so that client does need to put array accessors in templates.
		reqParams[k] = v[0]
	}

	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", pathParams["instance_id"]),
		attribute.String("args.name", pathParams["name"]),
	)

	s.addInstanceRequestAttributes(ctx, pathParams["instance_id"])

	api, err := s.runtime.APIForName(ctx, pathParams["instance_id"], pathParams["name"], reqParams)
	if err != nil {
		if errors.Is(err, runtime.ErrAPINotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Todo : for testing purposes only
	res := []byte(api.Spec.Sql)
	// TODO : what user attributes to get from ctx ?
	// this all will go in resolver.go in runtime may be

	// var resolverInitializer APIResolverInitializer
	// var ok bool
	// if api.Spec.Sql != "" {
	// 	resolverInitializer, ok = APIResolverInitializers["SQLResolver"]
	// 	if !ok {
	// 		panic("no SQLResolver")
	// 	}
	// } else {
	// 	resolverInitializer, ok = APIResolverInitializers["MetricsSQLResolver"]
	// 	if !ok {
	// 		return nil, status.Error(codes.InvalidArgument, "MetricsSQLResolver not found")
	// 	}
	// }

	// resolver, err := resolverInitializer(ctx, &APIResolverOptions{
	// 	Runtime:    r,
	// 	InstanceID: instanceID,
	// 	API:        api,
	// 	Args:       reqParams,
	// 	UserAttributes: reqParams,
	// })
	// if err != nil {
	// 	return nil, status.Error(codes.InvalidArgument, err.Error())
	// }

	// res, err := resolver.ResolveInteractive(ctx, 100)
	// if err != nil {
	// 	return nil, err
	// }

	// return res, nil

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(res)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to write response data: %s", err), http.StatusInternalServerError)
		return
	}
}
