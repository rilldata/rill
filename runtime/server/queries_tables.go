package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/queries"
	"google.golang.org/protobuf/types/known/structpb"
)

// Table level profiling APIs.
func (s *Server) GetTableCardinality(ctx context.Context, req *runtimev1.GetTableCardinalityRequest) (*runtimev1.GetTableCardinalityResponse, error) {
	q := &queries.TableCardinality{
		TableName: req.TableName,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}
	return &runtimev1.GetTableCardinalityResponse{
		Cardinality: q.Result,
	}, nil
}

type ColumnInfo struct {
	Name    string
	Type    string
	Unknown int
}

func EscapeDoubleQuotes(column string) string {
	return strings.ReplaceAll(column, "\"", "\"\"")
}

func EscapeHyphen(column string) string {
	return strings.ReplaceAll(column, "-", "_")
}

func (s *Server) ProfileColumns(ctx context.Context, req *runtimev1.ProfileColumnsRequest) (*runtimev1.ProfileColumnsResponse, error) {
	temporaryTableName := "profile_columns_" + EscapeHyphen(uuid.New().String())
	// views return duplicate column names, so we need to create a temporary table
	rows, err := s.query(ctx, req.InstanceId, &drivers.Statement{
		Query:    fmt.Sprintf(`CREATE TEMPORARY TABLE %q AS (SELECT * FROM %q LIMIT 1)`, temporaryTableName, req.TableName),
		Priority: int(req.Priority),
	})
	if err != nil {
		return nil, err
	}
	rows.Close()
	defer s.dropTempTable(req.InstanceId, int(req.Priority), temporaryTableName)

	rows, err = s.query(ctx, req.InstanceId, &drivers.Statement{
		Query: fmt.Sprintf(`select column_name as name, data_type as type from information_schema.columns 
		where table_name = '%s' and table_schema = 'temp'`, temporaryTableName),
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

	// Disabling this for now. we need to move this to a separate API
	// It adds a lot of response time to getting columns
	// for _, pc := range pcs[0:i] {
	//	columnName := EscapeDoubleQuotes(pc.Name)
	//	rows, err = s.query(ctx, req.InstanceId, &drivers.Statement{
	//		Query:    fmt.Sprintf(`select max(length("%s")) as max from %s`, columnName, req.TableName),
	//		Priority: int(req.Priority),
	//	})
	//	if err != nil {
	//		return nil, err
	//	}
	//	for rows.Next() {
	//		var max sql.NullInt32
	//		if err := rows.Scan(&max); err != nil {
	//			return nil, err
	//		}
	//		if max.Valid {
	//			pc.LargestStringLength = int32(max.Int32)
	//		}
	//	}
	//	rows.Close()
	//}

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
