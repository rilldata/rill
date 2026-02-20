package ai

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/template"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/jsonval"
)

// SchemaField represents a field a result's schema.
// Used to reduce result size through resolverResultToTabular.
type SchemaField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// resolverResultToTabular converts a resolver result into a tabular schema and row set.
func resolverResultToTabular(res runtime.ResolverResult) ([]SchemaField, [][]any, error) {
	// Build schema fields
	schema := res.Schema()
	fields := make([]SchemaField, len(schema.Fields))
	for i, f := range schema.Fields {
		fields[i] = SchemaField{
			Name: f.Name,
			Type: strings.TrimPrefix(f.Type.Code.String(), "CODE_"),
		}
	}

	// Collect rows as value slices in schema field order
	schemaType := &runtimev1.Type{Code: runtimev1.Type_CODE_STRUCT, StructType: schema}
	var data [][]any
	for {
		row, err := res.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, nil, err
		}

		v, err := jsonval.ToValue(row, schemaType)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to convert row: %w", err)
		}
		rowMap, ok := v.(map[string]any)
		if !ok {
			return nil, nil, fmt.Errorf("expected row to be map[string]any, got %T", v)
		}

		vals := make([]any, len(fields))
		for i, f := range fields {
			vals[i] = rowMap[f.Name]
		}
		data = append(data, vals)
	}

	return fields, data, nil
}

var templateFuncs = template.FuncMap{
	"backticks": func() string {
		return "```"
	},
}

func executeTemplate(templ string, data map[string]any) (string, error) {
	tmpl, err := template.New("").Funcs(templateFuncs).Parse(templ)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}

func mustExecuteTemplate(templ string, data map[string]any) string {
	result, err := executeTemplate(templ, data)
	if err != nil {
		panic(err)
	}
	return result
}
