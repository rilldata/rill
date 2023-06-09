package queries

import (
	"bytes"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	"testing"
)

func Test_writeCSV(t *testing.T) {
	meta := []*runtimev1.MetricsViewColumn{
		{
			Name: "col1\"",
		},
		{
			Name: "col2",
		},
	}
	fields := make(map[string]*structpb.Value)
	fields["col1\""] = structpb.NewStringValue("test\"doublequotes")
	fields["col2"] = structpb.NewStringValue("")

	data := []*structpb.Struct{
		{
			Fields: fields,
		},
	}

	var buf bytes.Buffer

	err := writeCSV(meta, data, &buf)
	require.NoError(t, err)

	expected := fmt.Sprintf(
		`"col1""",col2
"test""doublequotes",
`,
	)
	require.Equal(t, expected, buf.String())
	buf.Reset()

	delete(fields, "col1\"")
	delete(fields, "col2")
	meta = []*runtimev1.MetricsViewColumn{
		{
			Name: "col1",
		},
	}
	fields["col1"] = structpb.NewNumberValue(2.5)
	err = writeCSV(meta, data, &buf)
	require.NoError(t, err)
	require.Equal(t, "col1\n2.5\n", buf.String())
	buf.Reset()

	l := &structpb.ListValue{
		Values: []*structpb.Value{
			structpb.NewNumberValue(2.5),
			structpb.NewBoolValue(true),
		},
	}
	fields["col1"] = structpb.NewListValue(l)
	err = writeCSV(meta, data, &buf)
	require.NoError(t, err)
	require.Equal(t, "col1\n[2.5 true]\n", buf.String())
}
