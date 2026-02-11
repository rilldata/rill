package printer

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestToModelPartitionRow_IncludesRetryFields(t *testing.T) {
	p := &runtimev1.ModelPartition{
		Key:        "partition-1",
		Data:       mustStructPB(map[string]any{"country": "US"}),
		ExecutedOn: timestamppb.Now(),
		ElapsedMs:  1250,
		Error:      "boom",
		RetryUsed:  2,
		RetryMax:   3,
	}

	row := toModelPartitionRow(p)

	require.Equal(t, "partition-1", row.Key)
	require.Equal(t, `{"country":"US"}`, row.DataJSON)
	require.Equal(t, "boom", row.Error)
	require.Equal(t, "1.25s", row.Elapsed)
	require.Equal(t, "2/3", row.Retries)
}

func mustStructPB(v map[string]any) *structpb.Struct {
	s, err := structpb.NewStruct(v)
	if err != nil {
		panic(err)
	}
	return s
}
