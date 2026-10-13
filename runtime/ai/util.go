package ai

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"text/template"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/jsonval"
)

// isRillUserAgent reports whether the session belongs to a first-party Rill client (the in-app chat sends "rill/<version>").
// External MCP clients have other user agents; several tools are only useful, or only work, in one of the two.
func isRillUserAgent(s *Session) bool {
	return strings.HasPrefix(s.CatalogSession().UserAgent, "rill")
}

// aiMessageOpenURL returns the URL that opens the given tool call in the Rill UI (e.g. a query or chart replayed into an Explore).
// It takes the instance's frontend URL and returns an empty string when there is none (e.g. in tests or during rollout).
func aiMessageOpenURL(frontendURL, sessionID, callID string) (string, error) {
	if frontendURL == "" {
		return "", nil
	}

	openURL, err := url.Parse(frontendURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse frontend URL %q: %w", frontendURL, err)
	}
	openURL.Path, err = url.JoinPath(openURL.Path, "-", "ai", sessionID, "message", callID, "-", "open")
	if err != nil {
		return "", fmt.Errorf("failed to join path: %w", err)
	}
	return openURL.String(), nil
}

// SchemaField represents a field a result's schema.
// Used to reduce result size through resolverResultToTabular.
type SchemaField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// resolverResultToTabular converts a resolver result into a tabular schema and row set (rows as value slices in schema order).
func resolverResultToTabular(res runtime.ResolverResult) ([]SchemaField, [][]any, error) {
	fields, rows, err := resolverResultToRows(res)
	if err != nil {
		return nil, nil, err
	}
	data := make([][]any, len(rows))
	for i, row := range rows {
		vals := make([]any, len(fields))
		for j, f := range fields {
			vals[j] = row[f.Name]
		}
		data[i] = vals
	}
	return fields, data, nil
}

// resolverResultToRows converts a resolver result into its schema and rows keyed by field name, with JSON-friendly values.
func resolverResultToRows(res runtime.ResolverResult) ([]SchemaField, []map[string]any, error) {
	schema := res.Schema()
	fields := make([]SchemaField, len(schema.Fields))
	for i, f := range schema.Fields {
		fields[i] = SchemaField{
			Name: f.Name,
			Type: strings.TrimPrefix(f.Type.Code.String(), "CODE_"),
		}
	}

	schemaType := &runtimev1.Type{Code: runtimev1.Type_CODE_STRUCT, StructType: schema}
	var rows []map[string]any
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
		rows = append(rows, rowMap)
	}

	return fields, rows, nil
}

// normalizeFilePath converts a model-provided file path to an absolute path within the project (e.g. "models/foo.sql" to "/models/foo.sql").
// It returns an error for paths with ".." segments, which could otherwise traverse outside the project directory.
func normalizeFilePath(p string) (string, error) {
	segments := strings.FieldsFunc(p, func(r rune) bool { return r == '/' || r == '\\' })
	for _, segment := range segments {
		if segment == ".." {
			return "", fmt.Errorf("invalid path %q: must not contain %q segments", p, "..")
		}
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p, nil
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

func boolPtr(b bool) *bool { return &b }
