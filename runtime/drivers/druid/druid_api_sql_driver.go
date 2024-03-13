package druid

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
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

func toStringArray(values []any) []string {
	s := make([]string, len(values))
	for i, v := range values {
		vv, ok := v.(string)
		if !ok {
			return nil
		}
		s[i] = vv
	}
	return s
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
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(resp.Body)

	var obj any
	err = dec.Decode(&obj)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}
	switch v := obj.(type) {
	case map[string]any:
		resp.Body.Close()
		return nil, fmt.Errorf("%v", obj)
	case []any:
		columns := toStringArray(v)
		err = dec.Decode(&obj)
		if err != nil {
			resp.Body.Close()
			return nil, err
		}

		types := toStringArray(obj.([]any))

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
			columns:      columns,
			types:        types,
			transformers: transformers,
		}
		return druidRows, nil
	default:
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected response: %v", obj)
	}
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
	columns      []string
	types        []string
	transformers []func(any) (any, error)
}

var _ driver.Rows = &druidRows{}

func (dr *druidRows) Columns() []string {
	return dr.columns
}

func (dr *druidRows) Close() error {
	return dr.closer.Close()
}

func (dr *druidRows) Next(dest []driver.Value) error {
	var a []any
	err := dr.dec.Decode(&a)
	if err != nil {
		return err
	}

	for i, v := range a {
		v, err := dr.transformers[i](v)
		if err != nil {
			return err
		}

		dest[i] = v
	}

	return nil
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
		ResultFormat:   "arrayLines",
		Parameters:     parameters,
		Context: DruidQueryContext{
			SQLQueryID: uuid.New().String(),
		},
	}
}

func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, fmt.Errorf("unsupported")
}

func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	return nil, fmt.Errorf("unsupported")
}
