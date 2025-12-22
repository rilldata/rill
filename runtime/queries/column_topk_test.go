package queries_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/clickhouse"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestColumnTopKAgainstClickHouse(t *testing.T) {
	testmode.Expensive(t)

	ctx := context.Background()
	clickHouseContainer, err := clickhouse.RunContainer(ctx,
		testcontainers.WithImage("clickhouse/clickhouse-server:latest"),
		clickhouse.WithUsername("clickhouse"),
		clickhouse.WithPassword("clickhouse"),
		clickhouse.WithConfigFile("../testruntime/testdata/clickhouse-config.xml"),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := clickHouseContainer.Terminate(ctx)
		require.NoError(t, err)
	})

	host, err := clickHouseContainer.Host(ctx)
	require.NoError(t, err)
	port, err := clickHouseContainer.MappedPort(ctx, "9000/tcp")
	require.NoError(t, err)

	t.Setenv("RILL_RUNTIME_TEST_OLAP_DRIVER", "clickhouse")
	t.Setenv("RILL_RUNTIME_TEST_OLAP_DSN", fmt.Sprintf("clickhouse://clickhouse:clickhouse@%v:%v", host, port.Port()))
	t.Run("TestColumnTopK", func(t *testing.T) { TestColumnTopK(t) })
	t.Run("TestColumnTopKList", func(t *testing.T) { TestColumnTopKList(t) })
	t.Run("TestColumnTopKStruct", func(t *testing.T) { TestColumnTopKStruct(t) })
}

func TestColumnTopK(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "test", `
		SELECT 'abc' AS col, 1 AS val, TIMESTAMP '2022-11-01 00:00:00' AS times, DATE '2007-04-01' AS dates
		UNION ALL 
		SELECT 'def' AS col, 5 AS val, TIMESTAMP '2022-11-02 00:00:00' AS times, DATE '2009-06-01' AS dates
		UNION ALL 
		SELECT 'abc' AS col, 3 AS val, TIMESTAMP '2022-11-03 00:00:00' AS times, DATE '2010-04-11' AS dates
		UNION ALL 
		SELECT null AS col, 1 AS val, TIMESTAMP '2022-11-03 00:00:00' AS times, DATE '2010-11-21' AS dates
		UNION ALL 
		SELECT '12' AS col, 1 AS val, TIMESTAMP '2022-11-03 00:00:00' AS times, DATE '2011-06-30' AS dates
	`)

	q := &queries.ColumnTopK{
		TableName:  "test",
		ColumnName: "col",
		Agg:        "count(*)",
		K:          50,
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Equal(t, 4, len(q.Result.Entries))
	require.Equal(t, "abc", q.Result.Entries[0].Value.GetStringValue())
	require.Equal(t, 2, int(q.Result.Entries[0].Count))
	require.Equal(t, "12", q.Result.Entries[1].Value.GetStringValue())
	require.Equal(t, 1, int(q.Result.Entries[1].Count))
	require.Equal(t, "def", q.Result.Entries[2].Value.GetStringValue())
	require.Equal(t, 1, int(q.Result.Entries[2].Count))
	require.Equal(t, structpb.NewNullValue(), q.Result.Entries[3].Value)
	require.Equal(t, 1, int(q.Result.Entries[3].Count))

	q = &queries.ColumnTopK{
		TableName:  "test",
		ColumnName: "col",
		Agg:        "sum(val)",
		K:          50,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Equal(t, 4, len(q.Result.Entries))
	require.Equal(t, "def", q.Result.Entries[0].Value.GetStringValue())
	require.Equal(t, 5, int(q.Result.Entries[0].Count))
	require.Equal(t, "abc", q.Result.Entries[1].Value.GetStringValue())
	require.Equal(t, 4, int(q.Result.Entries[1].Count))
	require.Equal(t, "12", q.Result.Entries[2].Value.GetStringValue())
	require.Equal(t, 1, int(q.Result.Entries[2].Count))
	require.Equal(t, structpb.NewNullValue(), q.Result.Entries[3].Value)
	require.Equal(t, 1, int(q.Result.Entries[3].Count))

	q = &queries.ColumnTopK{
		TableName:  "test",
		ColumnName: "col",
		Agg:        "count(*)",
		K:          1,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Equal(t, 1, len(q.Result.Entries))
	require.Equal(t, "abc", q.Result.Entries[0].Value.GetStringValue())
	require.Equal(t, 2, int(q.Result.Entries[0].Count))
}

func TestColumnTopKList(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "test", `SELECT [10, 20] AS col, 1 AS val`)
	q := &queries.ColumnTopK{
		TableName:  "test",
		ColumnName: "col",
		Agg:        "count(*)",
		K:          10,
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Equal(t, 1, len(q.Result.Entries))
	require.Equal(t, 10.0, q.Result.Entries[0].Value.GetListValue().Values[0].GetNumberValue())
	require.Equal(t, 20.0, q.Result.Entries[0].Value.GetListValue().Values[1].GetNumberValue())
	require.Equal(t, 1, int(q.Result.Entries[0].Count))
}

func TestColumnTopKStruct(t *testing.T) {
	var rt *runtime.Runtime
	var instanceID string
	if os.Getenv("RILL_RUNTIME_TEST_OLAP_DRIVER") == "clickhouse" {
		rt, instanceID = testruntime.NewInstanceWithModel(t, "test", `SELECT CAST((['x', 'y'], [10, null]), 'Map(String,Nullable(UInt8))') AS col, 1 AS val`)
	} else {
		rt, instanceID = testruntime.NewInstanceWithModel(t, "test", `SELECT {'x': 10, 'y': null} AS col, 1 AS val`)
	}
	q := &queries.ColumnTopK{
		TableName:  "test",
		ColumnName: "col",
		Agg:        "count(*)",
		K:          10,
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Equal(t, 1, len(q.Result.Entries))
	require.Equal(t, 10.0, q.Result.Entries[0].Value.GetStructValue().AsMap()["x"])
	require.Equal(t, nil, q.Result.Entries[0].Value.GetStructValue().AsMap()["y"])
	require.Equal(t, 1, int(q.Result.Entries[0].Count))
}

func TestColumnTopKJSON(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "test", `SELECT '{"a": 10}'::JSON AS col, 1 AS val`)
	q := &queries.ColumnTopK{
		TableName:  "test",
		ColumnName: "col",
		Agg:        "count(*)",
		K:          10,
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Equal(t, 1, len(q.Result.Entries))
	require.Equal(t, 10.0, q.Result.Entries[0].Value.GetStructValue().AsMap()["a"])
	require.Equal(t, 1, int(q.Result.Entries[0].Count))
}

func BenchmarkColumnTopK(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := &queries.ColumnTopK{
			TableName:  "ad_bids",
			ColumnName: "domain",
			Agg:        "sum(bid_price)",
			K:          50,
		}
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}
