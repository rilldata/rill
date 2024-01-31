package server

import (
	"context"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"go.opentelemetry.io/otel/attribute"
)

func (s *Server) Telemetry(ctx context.Context, req *adminv1.TelemetryRequest) (*adminv1.TelemetryResponse, error) {
	dims := make([]attribute.KeyValue, 0)
	for k, v := range req.Event.AsMap() {
		a, ok := toKeyValue(k, v)
		if ok {
			dims = append(dims, a)
		}
	}
	s.uiActivity.Emit(ctx, req.Name, float64(req.Value), dims...)
	return &adminv1.TelemetryResponse{}, nil
}

func toKeyValue(k string, v interface{}) (attribute.KeyValue, bool) {
	switch vt := v.(type) {
	case string:
		return attribute.String(k, vt), true
	case bool:
		return attribute.Bool(k, vt), true
	case int:
		return attribute.Int(k, vt), true
	case int64:
		return attribute.Int64(k, vt), true
	case float64:
		return attribute.Float64(k, vt), true
	}

	return attribute.KeyValue{}, false
}
