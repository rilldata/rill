package server

import (
	"context"
	"fmt"
	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const DefaultK = 50
const DefaultAgg = "count(*)"

func (s *Server) GetTopK(ctx context.Context, topKRequest *api.TopKRequest) (*api.TopKResponse, error) {
	agg := DefaultAgg
	k := int32(DefaultK)
	if topKRequest.Agg != nil {
		agg = *topKRequest.Agg
	}
	if topKRequest.K != nil {
		k = *topKRequest.K
	}
	topKSql := fmt.Sprintf("SELECT %s as value, %s AS count from %s GROUP BY %s ORDER BY count desc LIMIT %d",
		QuoteName(topKRequest.ColumnName),
		agg,
		topKRequest.TableName,
		QuoteName(topKRequest.ColumnName),
		k,
	)
	rows, err := s.query(ctx, topKRequest.InstanceId, &drivers.Statement{
		Query: topKSql,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	meta, err := rowsToMeta(rows)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	data, err := rowsToData(rows)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &api.TopKResponse{
		Meta: meta,
		Data: data,
	}
	return resp, nil
}

func QuoteName(columnName string) string {
	return fmt.Sprintf("\"%s\"", columnName)
}
