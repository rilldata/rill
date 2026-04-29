package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	jsonvalue "github.com/Andrew-M-C/go.jsonvalue"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

var mutex sync.Mutex

func unmarshalJSON(sql []byte) (*jsonvalue.V, error) {
	mutex.Lock()
	defer mutex.Unlock()

	return jsonvalue.Unmarshal(sql)
}

// Query implements QueryService.
func (s *Server) Query(ctx context.Context, req *runtimev1.QueryRequest) (*runtimev1.QueryResponse, error) {
	if !auth.GetClaims(ctx, req.InstanceId).Can(runtime.ReadOLAP) {
		return nil, ErrForbidden
	}

	args := make([]any, len(req.Args))
	for i, arg := range req.Args {
		args[i] = arg.AsInterface()
	}

	olap, release, err := s.runtime.OLAP(ctx, req.InstanceId, req.Connector)
	if err != nil {
		return nil, err
	}
	defer release()

	transformedSQL := req.Sql
	if req.Limit != 0 {
		transformedSQL, err = ensureLimits(ctx, olap, req.Sql, int(req.Limit))
		if err != nil {
			return nil, err
		}
	}

	res, err := olap.Query(ctx, &drivers.Statement{
		Query:            transformedSQL,
		Args:             args,
		DryRun:           req.DryRun,
		Priority:         int(req.Priority),
		ExecutionTimeout: 2 * time.Minute,
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
		return nil, err
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

		rowStruct, err := pbutil.ToStruct(rowMap, rows.Schema)
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

func ensureLimits(ctx context.Context, olap drivers.OLAPStore, inputSQL string, limit int) (string, error) {
	r, err := olap.Query(ctx, &drivers.Statement{
		Query: "select json_serialize_sql(?::VARCHAR)::BLOB",
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

	v, err := unmarshalJSON(serializedSQL)
	if err != nil {
		return "", err
	}

	err = transformStatments(v, limit)
	if err != nil {
		return "", err
	}

	transformedJSON := v.MustMarshalString()
	r, err = olap.Query(ctx, &drivers.Statement{
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

func transformStatments(root *jsonvalue.V, limit int) error {
	for _, v := range root.MustGet("statements").ForRangeArr() {
		err := replaceOrUpdateLimitTo(v.MustGet("node").MustGet("modifiers"), limit)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
"LIMIT 1" clause is serialized to the following JSON:

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

"LIMIT ?" clause serialization is:

	{
	   "type":"LIMIT_MODIFIER",
	   "limit":{
	      "class":"PARAMETER",
	      "type":"VALUE_PARAMETER",
	      "alias":"",
	      "parameter_nr":2
	   },
	   "offset":null
	}
*/
func replaceOrUpdateLimitTo(root *jsonvalue.V, limit int) error {
	children := root.ForRangeArr()
	updated := false
	if len(children) != 0 {
		for _, v := range children {
			if v.MustGet("type").String() == "LIMIT_MODIFIER" {
				modifierType := v.MustGet("limit").MustGet("class").String()
				switch modifierType {
				case "CONSTANT":
					v.MustGet("limit").MustGet("value").MustSetInt(limit).At("value")
					updated = true
				case "PARAMETER":
					err := v.Delete("limit")
					if err != nil {
						return err
					}

					limitObject, err := createConstantLimit(limit)
					if err != nil {
						return err
					}

					v.MustSet(limitObject).At("limit")
					updated = true
				}
			}
		}
	}

	if !updated {
		v, err := createLimitModifier(limit)
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

func createConstantLimit(limit int) (*jsonvalue.V, error) {
	return unmarshalJSON([]byte(fmt.Sprintf(`
	{
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
	}
`, limit)))
}

func createLimitModifier(limit int) (*jsonvalue.V, error) {
	return unmarshalJSON([]byte(fmt.Sprintf(`
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
}
