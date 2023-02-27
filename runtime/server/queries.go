package server

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

// Query implements RuntimeService.
func (s *Server) Query(ctx context.Context, req *runtimev1.QueryRequest) (*runtimev1.QueryResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadOLAP) {
		return nil, ErrForbidden
	}

	args := make([]any, len(req.Args))
	for i, arg := range req.Args {
		args[i] = arg.AsInterface()
	}

	olap, err := s.runtime.OLAP(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	res, err := olap.Execute(ctx, &drivers.Statement{
		Query:    req.Sql,
		Args:     args,
		DryRun:   req.DryRun,
		Priority: int(req.Priority),
	})
	if err != nil {
		// TODO: Parse error to determine error code
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// NOTE: Currently, query returns nil res for successful dry-run queries
	if req.DryRun {
		// TODO: Return a meta object for dry-run queries
		return &runtimev1.QueryResponse{}, nil
	}

	defer res.Close()

	data, err := rowsToData(res)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &runtimev1.QueryResponse{
		Meta: res.Schema,
		Data: data,
	}

	return resp, nil
}

func rowsToData(rows *drivers.Result) ([]*structpb.Struct, error) {
	var data []*structpb.Struct
	for rows.Next() {
		rowMap := make(map[string]any)
		err := rows.MapScan(rowMap)
		if err != nil {
			return nil, err
		}

		rowStruct, err := pbutil.ToStruct(rowMap)
		if err != nil {
			return nil, err
		}

		data = append(data, rowStruct)
	}

	err := rows.Err()
	if err != nil {
		return nil, err
	}

	return data, nil
}
