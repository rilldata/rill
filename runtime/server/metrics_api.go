package server

import (
	"context"
	"github.com/rilldata/rill/runtime/api"
)

func (s *Server) GenerateTimeSeries(ctx context.Context, in *api.GenerateTimeSeriesRequest) (*api.TimeSeriesRollup, error) {
	return &api.TimeSeriesRollup{
		Rollup: &api.TimeSeriesResponse{},
	}, nil
}
