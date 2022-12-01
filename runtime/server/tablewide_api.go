package server

import (
	"context"
	"fmt"
	"regexp"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/structpb"
)

// Table level profiling APIs
func (s *Server) GetTableCardinality(ctx context.Context, req *runtimev1.GetTableCardinalityRequest) (*runtimev1.GetTableCardinalityResponse, error) {
	rows, err := s.query(ctx, req.InstanceId, &drivers.Statement{
		Query:    "select count(*) from " + quoteName(req.TableName),
		Priority: int(req.Priority),
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var count int64
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return nil, err
		}
	}
	return &runtimev1.GetTableCardinalityResponse{
		Cardinality: count,
	}, nil
}

type ColumnInfo struct {
	Name    string
	Type    string
	Unknown int
}

var DoubleQuotesRegexp *regexp.Regexp = regexp.MustCompile("\"")

func EscapeDoubleQuotes(column string) string {
	return DoubleQuotesRegexp.ReplaceAllString(column, "\"\"")
}

func (s *Server) ProfileColumns(ctx context.Context, req *runtimev1.ProfileColumnsRequest) (*runtimev1.ProfileColumnsResponse, error) {
	rows, err := s.query(ctx, req.InstanceId, &drivers.Statement{
		Query: fmt.Sprintf(`select column_name as name, data_type as type from information_schema.columns 
		where table_name = '%s' and table_schema = current_schema()`, req.TableName),
		Priority: int(req.Priority),
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var pcs []*runtimev1.ProfileColumn
	i := 0
	for rows.Next() {
		pc := runtimev1.ProfileColumn{}
		if err := rows.StructScan(&pc); err != nil {
			return nil, err
		}
		pcs = append(pcs, &pc)
		i++
	}

	for _, pc := range pcs[0:i] {
		columnName := EscapeDoubleQuotes(pc.Name)
		rows, err = s.query(ctx, req.InstanceId, &drivers.Statement{
			Query:    fmt.Sprintf(`select max(length("%s")) as max from %s`, columnName, req.TableName),
			Priority: int(req.Priority),
		})
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			if err := rows.Scan(&pc.LargestStringLength); err != nil {
				return nil, err
			}
		}
		rows.Close()
	}

	return &runtimev1.ProfileColumnsResponse{
		ProfileColumns: pcs[0:i],
	}, nil
}

func (s *Server) GetTableRows(ctx context.Context, req *runtimev1.GetTableRowsRequest) (*runtimev1.GetTableRowsResponse, error) {
	rows, err := s.query(ctx, req.InstanceId, &drivers.Statement{
		Query:    fmt.Sprintf("select * from %s limit %d", req.TableName, req.Limit),
		Priority: int(req.Priority),
	})
	if err != nil {
		return nil, err
	}
	var data []*structpb.Struct
	if data, err = rowsToData(rows); err != nil {
		return nil, err
	}

	return &runtimev1.GetTableRowsResponse{
		Data: data,
	}, nil
}
