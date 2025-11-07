package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
)

type TableColumns struct {
	Connector      string
	Database       string
	DatabaseSchema string
	TableName      string
	Result         *runtimev1.TableColumnsResponse
}

var _ runtime.Query = &TableColumns{}

func (q *TableColumns) Key() string {
	return fmt.Sprintf("TableColumns:%s", q.TableName)
}

func (q *TableColumns) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindSource, Name: q.TableName},
		{Kind: runtime.ResourceKindModel, Name: q.TableName},
	}
}

func (q *TableColumns) MarshalResult() *runtime.QueryResult {
	var size int64
	if len(q.Result.ProfileColumns) > 0 {
		// approx
		size = sizeProtoMessage(q.Result.ProfileColumns[0]) * int64(len(q.Result.ProfileColumns))
	}
	if len(q.Result.UnsupportedColumns) > 0 {
		r, err := json.Marshal(q.Result.UnsupportedColumns)
		if err == nil { // ignore error
			size += int64(len(r))
		}
	}
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: size,
	}
}

func (q *TableColumns) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.TableColumnsResponse)
	if !ok {
		return fmt.Errorf("TableColumns: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *TableColumns) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, release, err := rt.OLAP(ctx, instanceID, q.Connector)
	if err != nil {
		return err
	}
	defer release()

	if !supportedTableHeadDialects[olap.Dialect()] {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	tbl, err := olap.InformationSchema().Lookup(ctx, q.Database, q.DatabaseSchema, q.TableName)
	if err != nil {
		return err
	}

	q.Result = &runtimev1.TableColumnsResponse{
		ProfileColumns:     make([]*runtimev1.ProfileColumn, len(tbl.Schema.Fields)),
		UnsupportedColumns: tbl.UnsupportedCols,
	}
	for i := 0; i < len(tbl.Schema.Fields); i++ {
		q.Result.ProfileColumns[i] = &runtimev1.ProfileColumn{
			Name: tbl.Schema.Fields[i].Name,
			Type: tbl.Schema.Fields[i].Type.Code.String(),
		}
	}
	return nil
}

func (q *TableColumns) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return ErrExportNotSupported
}
