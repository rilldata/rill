package druid

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type druidSQLDriver struct{}

var _ driver.Driver = &druidSQLDriver{}

func (a *druidSQLDriver) Open(dsn string) (driver.Conn, error) {
	client := http.Client{Timeout: time.Second * 10}

	return &sqlConnection{
		client: &client,
		dsn:    dsn,
	}, nil
}

func init() {
	sql.Register("druid", &druidSQLDriver{})
}

type sqlConnection struct {
	client *http.Client
	dsn    string
}

var _ driver.QueryerContext = &sqlConnection{}

func emptyTransformer(v any) (any, error) {
	return v, nil
}

func (c *sqlConnection) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	b, err := json.Marshal(druidRequest(query, args))
	if err != nil {
		return nil, err
	}

	bodyReader := bytes.NewReader(b)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.dsn, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	// nolint:all
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(resp.Body)

	jw := &JSONWalker{
		dec: dec,
	}

	// nolint:all
	if !jw.enterArrayOrError() {
		return nil, jw.err
	}

	if jw.enterArray() {
		columns, err := jw.stringArrayValues()
		if err != nil {
			return nil, err
		}
		if jw.enterArray() {
			types, err := jw.stringArrayValues()
			if err != nil {
				return nil, err
			}

			transformers := make([]func(any) (any, error), len(columns))
			for i, c := range types {
				transformers[i] = emptyTransformer
				if c == "TIMESTAMP" {
					transformers[i] = func(v any) (any, error) {
						t, err := time.Parse(time.RFC3339, v.(string))
						if err != nil {
							return nil, err
						}
						return t, nil
					}
				} else if c == "ARRAY" {
					transformers[i] = func(v any) (any, error) {
						var l []any
						err := json.Unmarshal([]byte(v.(string)), &l)
						if err != nil {
							return nil, err
						}
						return l, nil
					}
				} else if c == "OTHER" {
					transformers[i] = func(v any) (any, error) {
						var l map[string]any
						err := json.Unmarshal([]byte(v.(string)), &l)
						if err != nil {
							return nil, err
						}
						return l, nil
					}
				}
			}

			druidRows := &druidRows{
				closer:       resp.Body,
				dec:          dec,
				jw:           jw,
				columns:      columns,
				types:        types,
				transformers: transformers,
			}
			return druidRows, nil
		}
	}
	return nil, jw.err
}

func toType(v any) string {
	switch v.(type) {
	case int:
		return "INTEGER"
	case float64:
		return "DOUBLE"
	case bool:
		return "BOOLEAN"
	default:
		return "VARCHAR"
	}
}

type druidRows struct {
	closer       io.ReadCloser
	dec          *json.Decoder
	jw           *JSONWalker
	columns      []string
	types        []string
	transformers []func(any) (any, error)
}

func (dr *druidRows) Columns() []string {
	return dr.columns
}

func (dr *druidRows) Close() error {
	return dr.closer.Close()
}

func (dr *druidRows) Next(dest []driver.Value) error {
	if !dr.jw.hasMore() {
		return io.EOF
	} else if dr.jw.enterArray() {
		values, err := dr.jw.arrayValues()
		if err != nil {
			return err
		}

		for i, v := range values {
			v, err := dr.transformers[i](v)
			if err != nil {
				return err
			}

			dest[i] = v
		}
		return nil
	}
	return dr.jw.err
}

type stmt struct {
	conn  *sqlConnection
	query string
}

func (c *sqlConnection) Prepare(query string) (driver.Stmt, error) {
	return &stmt{
		query: query,
		conn:  c,
	}, nil
}

func (c *sqlConnection) Close() error {
	return nil
}

func (c *sqlConnection) Begin() (driver.Tx, error) {
	return nil, fmt.Errorf("unsupported")
}

func (s *stmt) Close() error {
	return nil
}

func (s *stmt) NumInput() int {
	return 0
}

type DruidQueryContext struct {
	SQLQueryID string `json:"sqlQueryId"`
}

type DruidParameter struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type DruidRequest struct {
	Query          string            `json:"query"`
	Header         bool              `json:"header"`
	SQLTypesHeader bool              `json:"sqlTypesHeader"`
	ResultFormat   string            `json:"resultFormat"`
	Parameters     []DruidParameter  `json:"parameters"`
	Context        DruidQueryContext `json:"context"`
}

func druidRequest(query string, args []driver.NamedValue) *DruidRequest {
	parameters := make([]DruidParameter, len(args))
	for i, arg := range args {
		parameters[i] = DruidParameter{
			Type:  toType(arg.Value),
			Value: arg.Value,
		}
	}
	return &DruidRequest{
		Query:          query,
		Header:         true,
		SQLTypesHeader: true,
		ResultFormat:   "array",
		Parameters:     parameters,
		Context: DruidQueryContext{
			SQLQueryID: uuid.New().String(),
		},
	}
}

func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, fmt.Errorf("unsupported")
}

type JSONWalker struct {
	dec *json.Decoder
	err error
}

func (j *JSONWalker) hasMore() bool {
	return j.dec.More()
}

func (j *JSONWalker) enterArrayOrError() bool {
	t, err := j.dec.Token()
	if err != nil {
		j.err = err
		return false
	}
	d, ok := t.(json.Delim)
	if ok && d == '[' {
		return true
	} else if d == '{' {
		bytes1 := make([]byte, 1024)
		n1, err := j.dec.Buffered().Read(bytes1)
		if err != nil {
			j.err = err
		}

		j.err = errors.New(string(bytes1[:n1]))
	}
	return false
}

func (j *JSONWalker) enterArray() bool {
	t, err := j.dec.Token()
	if err != nil {
		j.err = err
		return false
	}
	if d, ok := t.(json.Delim); ok && d == '[' {
		return true
	}
	var b []byte
	_, _ = j.dec.Buffered().Read(b)
	j.err = fmt.Errorf("expected array: %v %v", t, string(b))
	return false
}

func (j *JSONWalker) exitArray() bool {
	t, err := j.dec.Token()

	if err != nil {
		j.err = err
		return false
	}
	if d, ok := t.(json.Delim); ok && d == ']' {
		return true
	}
	var b []byte
	_, _ = j.dec.Buffered().Read(b)
	j.err = fmt.Errorf("expected array end: %v %v", t, string(b))

	return false
}

func (j *JSONWalker) arrayValues() ([]any, error) {
	var values []any
	for j.dec.More() {
		t, err := j.dec.Token()

		if err != nil {
			return nil, err
		}
		values = append(values, t)
	}
	if !j.exitArray() {
		return nil, j.err
	}
	return values, nil
}

func (j *JSONWalker) stringArrayValues() ([]string, error) {
	var columns []string
	for j.dec.More() {
		t, err := j.dec.Token()

		if err != nil {
			return nil, err
		}
		if s, ok := t.(string); ok {
			columns = append(columns, s)
		}
	}
	if !j.exitArray() {
		return nil, j.err
	}
	return columns, nil
}

func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	return nil, fmt.Errorf("unsupported")
}
