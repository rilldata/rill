package queries

import (
	"bytes"
	"fmt"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/types/known/structpb"
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

	err := WriteCSV(meta, data, &buf)
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

	err := WriteCSV(meta, data, &buf)
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

	err := WriteCSV(meta, data, &buf)
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

	err := WriteCSV(meta, data, &buf)
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

	err := WriteCSV(meta, data, &buf)
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
	err := WriteCSV(meta, data, &buf)
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

	err := WriteCSV(meta, data, &buf)
	require.NoError(t, err)

	expected := fmt.Sprintf(
		`"col1""",col2
"test""doublequotes",
`,
	)
	require.Equal(t, expected, buf.String())
	buf.Reset()
}

func Test_writeXLSX_emptystring(t *testing.T) {
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

	err := WriteXLSX(meta, data, &buf)
	require.NoError(t, err)

	file, err := excelize.OpenReader(&buf)
	require.NoError(t, err)

	v, err := file.GetCellValue("Sheet1", "A1")
	require.NoError(t, err)
	require.Equal(t, "col", v)

	v, err = file.GetCellValue("Sheet1", "A2")
	require.NoError(t, err)
	require.Equal(t, "", v)
}

func Test_writeXLSX_size(t *testing.T) {
	meta := []*runtimev1.MetricsViewColumn{
		{
			Name: "col",
		},
	}
	fields := make(map[string]*structpb.Value)
	fields["col"] = structpb.NewNumberValue(1)
	data := []*structpb.Struct{
		{
			Fields: fields,
		},
		{
			Fields: fields,
		},
	}

	var buf bytes.Buffer

	err := WriteXLSX(meta, data, &buf)
	require.NoError(t, err)

	file, err := excelize.OpenReader(&buf)
	require.NoError(t, err)

	/*
		should be
		|   ||A   |B  |
		|===||====|===|
		|1  ||col |   |
		|2  ||1   |   |
		|3  ||1   |   |
		|4  ||    |   |
	*/

	v, err := file.GetCellValue("Sheet1", "A1")
	require.NoError(t, err)
	require.Equal(t, "col", v)

	v, err = file.GetCellValue("Sheet1", "B1")
	require.NoError(t, err)
	require.Equal(t, "", v)

	v, err = file.GetCellValue("Sheet1", "A2")
	require.NoError(t, err)
	require.Equal(t, "1", v)

	v, err = file.GetCellValue("Sheet1", "B2")
	require.NoError(t, err)
	require.Equal(t, "", v)

	v, err = file.GetCellValue("Sheet1", "A3")
	require.NoError(t, err)
	require.Equal(t, "1", v)

	v, err = file.GetCellValue("Sheet1", "B3")
	require.NoError(t, err)
	require.Equal(t, "", v)

	v, err = file.GetCellValue("Sheet1", "A4")
	require.NoError(t, err)
	require.Equal(t, "", v)

	v, err = file.GetCellValue("Sheet1", "B4")
	require.NoError(t, err)
	require.Equal(t, "", v)
}

func Test_writeXLSX_number(t *testing.T) {
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

	err := WriteXLSX(meta, data, &buf)
	require.NoError(t, err)

	file, err := excelize.OpenReader(&buf)
	require.NoError(t, err)

	v, err := file.GetCellValue("Sheet1", "A2")
	require.NoError(t, err)
	require.Equal(t, "2.5", v)
}

func Test_writeXLSX_null(t *testing.T) {
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

	err := WriteXLSX(meta, data, &buf)
	require.NoError(t, err)

	file, err := excelize.OpenReader(&buf)
	require.NoError(t, err)

	v, err := file.GetCellValue("Sheet1", "A2")
	require.NoError(t, err)
	require.Equal(t, "", v)
}

func Test_writeXLSX_bool(t *testing.T) {
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

	err := WriteXLSX(meta, data, &buf)
	require.NoError(t, err)

	file, err := excelize.OpenReader(&buf)
	require.NoError(t, err)

	v, err := file.GetCellValue("Sheet1", "A2")
	require.NoError(t, err)
	require.Equal(t, "TRUE", v)
}

func Test_writeXLSX_struct(t *testing.T) {
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

	err := WriteXLSX(meta, data, &buf)
	require.NoError(t, err)

	file, err := excelize.OpenReader(&buf)
	require.NoError(t, err)

	v, err := file.GetCellValue("Sheet1", "A2")
	require.NoError(t, err)
	require.Equal(t, "{\"a\":2.5}", v)
}

func Test_writeXLSX_list(t *testing.T) {
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

	err := WriteXLSX(meta, data, &buf)
	require.NoError(t, err)

	file, err := excelize.OpenReader(&buf)
	require.NoError(t, err)

	v, err := file.GetCellValue("Sheet1", "A2")
	require.NoError(t, err)
	require.Equal(t, "[2.5 true]", v)
}

func Test_writeXLSX_quotes(t *testing.T) {
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
	fields["col"] = structpb.NewStringValue("a\"")

	var buf bytes.Buffer

	err := WriteXLSX(meta, data, &buf)
	require.NoError(t, err)

	file, err := excelize.OpenReader(&buf)
	require.NoError(t, err)

	v, err := file.GetCellValue("Sheet1", "A2")
	require.NoError(t, err)
	require.Equal(t, "a\"", v)
}
