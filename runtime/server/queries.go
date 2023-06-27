package server

import (
	"context"
	"fmt"
	"time"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

// Query implements QueryService.
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

func (s *Server) CustomQuery(ctx context.Context, req *runtimev1.CustomQueryRequest) (*runtimev1.CustomQueryResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadOLAP) {
		return nil, ErrForbidden
	}

	olap, err := s.runtime.OLAP(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	transformedSQL, err := ensureLimits(ctx, olap, req.Sql)
	if err != nil {
		return nil, err
	}

	res, err := olap.Execute(ctx, &drivers.Statement{
		Query:            transformedSQL,
		Priority:         int(req.Priority),
		ExecutionTimeout: 2 * time.Minute,
	})
	if err != nil {
		return nil, err
	}

	defer res.Close()

	data, err := rowsToData(res)
	if err != nil {
		return nil, err
	}

	resp := &runtimev1.CustomQueryResponse{
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

func ensureLimits(ctx context.Context, olap drivers.OLAPStore, inputSQL string) (string, error) {
	r, err := olap.Execute(ctx, &drivers.Statement{
		Query: "select json_serialize_sql(?::VARCHAR)",
		Args:  []any{inputSQL},
	})
	if err != nil {
		return "", err
	}

	var serializedSQL []byte
	if r.Next() {
		err = r.Scan(&serializedSQL)
		if err != nil {
			r.Close()
			return "", err
		}
	}

	r.Close()

	v, err := jsonvalue.Unmarshal(serializedSQL)
	if err != nil {
		return "", err
	}

	err = traverseAndUpdateModifiers(v)
	if err != nil {
		return "", err
	}

	transformedJSON := v.MustMarshalString()
	r, err = olap.Execute(ctx, &drivers.Statement{
		Query: "select json_deserialize_sql(json(?))",
		Args:  []any{transformedJSON},
	})
	if err != nil {
		return "", err
	}

	var sqlString string
	if r.Next() {
		err = r.Scan(&sqlString)
		if err != nil {
			r.Close()
			return "", err
		}
	}

	r.Close()

	return sqlString, nil
}

func traverseAndUpdateModifiers(root *jsonvalue.V) error {
	if root.IsArray() {
		for _, v := range root.ForRangeArr() {
			err := traverseAndUpdateModifiers(v)
			if err != nil {
				return err
			}
		}
	} else if root.IsObject() {
		for k, v := range root.ForRangeObj() {
			if k == "modifiers" {
				err := replaceOrUpdateLimitTo(v, 100)
				if err != nil {
					return err
				}
			} else {
				err := traverseAndUpdateModifiers(v)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

/*
LIMIT claused is serialized to the following JSON:

"modifiers":[

	{
	   "type":"LIMIT_MODIFIER",
	   "limit":{
	      "class":"CONSTANT",
	      "type":"VALUE_CONSTANT",
	      "alias":"",
	      "value":{
	         "type":{
	            "id":"INTEGER",
	            "type_info":null
	         },
	         "is_null":false,
	         "value":1
	      }
	   },
	   "offset":null
	},

]
*/
func replaceOrUpdateLimitTo(root *jsonvalue.V, limit int) error {
	children := root.ForRangeArr()
	updated := false
	if len(children) != 0 {
		for _, v := range children {
			if v.MustGet("type").String() == "LIMIT_MODIFIER" && v.MustGet("limit").MustGet("class").String() == "CONSTANT" {
				v.MustGet("limit").MustGet("value").MustSetInt(limit).At("value")
				updated = true
			}
		}
	}

	if !updated {
		v, err := createLimit(limit)
		if err != nil {
			return err
		}

		_, err = root.Append(v).InTheEnd()
		if err != nil {
			return err
		}
	}

	return nil
}

func createLimit(limit int) (*jsonvalue.V, error) {
	v, err := jsonvalue.Unmarshal([]byte(fmt.Sprintf(`
{
	"type":"LIMIT_MODIFIER",
	"limit":{
	   "class":"CONSTANT",
	   "type":"VALUE_CONSTANT",
	   "alias":"",
	   "value":{
		  "type":{
			 "id":"INTEGER",
			 "type_info":null
		  },
		  "is_null":false,
		  "value":%d
	   }
	},
	"offset":null
 }
`, limit)))
	return v, err
}
