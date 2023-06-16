package queries

import (
	"bytes"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	"testing"
)

func Test_writeCSV_emptystring(t *testing.T) {
	meta := []*runtimev1.MetricsViewColumn{
		{
			Name: "col",
		},
	}
	fields := make(map[string]*structpb.Value)
	fields["col"] = structpb.NewStringValue("")
	data := []*structpb.Struct{
		{
			Fields: fields,
		},
	}

	var buf bytes.Buffer

	err := writeCSV(meta, data, &buf)
	require.NoError(t, err)
	require.Equal(t, "col\n\n", buf.String())
}

func Test_writeCSV_number(t *testing.T) {
	meta := []*runtimev1.MetricsViewColumn{
		{
			Name: "col",
		},
	}
	fields := make(map[string]*structpb.Value)
	fields["col"] = structpb.NewNumberValue(2.5)
	data := []*structpb.Struct{
		{
			Fields: fields,
		},
	}

	var buf bytes.Buffer

	err := writeCSV(meta, data, &buf)
	require.NoError(t, err)
	require.Equal(t, "col\n2.5\n", buf.String())
}

func Test_writeCSV_null(t *testing.T) {
	meta := []*runtimev1.MetricsViewColumn{
		{
			Name: "col",
		},
	}
	fields := make(map[string]*structpb.Value)
	fields["col"] = structpb.NewNullValue()
	data := []*structpb.Struct{
		{
			Fields: fields,
		},
	}

	var buf bytes.Buffer

	err := writeCSV(meta, data, &buf)
	require.NoError(t, err)
	require.Equal(t, "col\n\n", buf.String())
}

func Test_writeCSV_bool(t *testing.T) {
	meta := []*runtimev1.MetricsViewColumn{
		{
			Name: "col",
		},
	}
	fields := make(map[string]*structpb.Value)
	fields["col"] = structpb.NewBoolValue(true)
	data := []*structpb.Struct{
		{
			Fields: fields,
		},
	}

	var buf bytes.Buffer

	err := writeCSV(meta, data, &buf)
	require.NoError(t, err)
	require.Equal(t, "col\ntrue\n", buf.String())
}

func Test_writeCSV_struct(t *testing.T) {
	meta := []*runtimev1.MetricsViewColumn{
		{
			Name: "col",
		},
	}
	fields := make(map[string]*structpb.Value)
	subfields := make(map[string]*structpb.Value)
	subfields["a"] = structpb.NewNumberValue(2.5)

	fields["col"] = structpb.NewStructValue(&structpb.Struct{
		Fields: subfields,
	})
	data := []*structpb.Struct{
		{
			Fields: fields,
		},
	}

	var buf bytes.Buffer

	err := writeCSV(meta, data, &buf)
	require.NoError(t, err)
	require.Equal(t, "col\n\"{\"\"a\"\":2.5}\"\n", buf.String())
}

func Test_writeCSV_list(t *testing.T) {
	meta := []*runtimev1.MetricsViewColumn{
		{
			Name: "col",
		},
	}
	fields := make(map[string]*structpb.Value)
	data := []*structpb.Struct{
		{
			Fields: fields,
		},
	}

	fields["col"] = structpb.NewListValue(
		&structpb.ListValue{
			Values: []*structpb.Value{
				structpb.NewNumberValue(2.5),
				structpb.NewBoolValue(true),
			},
		},
	)

	var buf bytes.Buffer
	err := writeCSV(meta, data, &buf)
	require.NoError(t, err)
	require.Equal(t, "col\n[2.5 true]\n", buf.String())
}

func Test_writeCSV_quotes(t *testing.T) {
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
}
