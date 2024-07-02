package druidsqldriver

import (
	"bytes"
	"context"
	"crypto/x509"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime/drivers/druid/retrier"
)

// non-retryable HTTP errors
var (
	redirectsErrorRe  = regexp.MustCompile(`stopped after \d+ redirects\z`)
	schemeErrorRe     = regexp.MustCompile(`unsupported protocol scheme`)
	notTrustedErrorRe = regexp.MustCompile(`certificate is not trusted`)
)

type druidSQLDriver struct{}

var _ driver.Driver = &druidSQLDriver{}

func (a *druidSQLDriver) Open(dsn string) (driver.Conn, error) {
	client := http.Client{}

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

type coordinatorHTTPCheck struct {
	c *sqlConnection
}

var _ retrier.AdditionalTest = &coordinatorHTTPCheck{}

func (chc *coordinatorHTTPCheck) EnsureHardFailure(ctx context.Context) (bool, error) {
	dr := newDruidRequest("SELECT * FROM sys.segments LIMIT 1", nil)
	b, err := json.Marshal(dr)
	if err != nil {
		return false, err
	}

	bodyReader := bytes.NewReader(b)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, chc.c.dsn, bodyReader)
	if err != nil {
		return false, err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := chc.c.client.Do(req)
	if err != nil {
		return false, err
	}

	dec := json.NewDecoder(resp.Body)

	var obj any
	err = dec.Decode(&obj)
	resp.Body.Close()
	if err != nil {
		return false, err
	}
	switch v := obj.(type) {
	case map[string]any:
		if v["errorCode"] != "invalidInput" {
			return false, fmt.Errorf("%v", obj)
		}
		return true, nil
	case []any:
		return true, nil
	default:
		return false, fmt.Errorf("unexpected response: %v", obj)
	}
}

func (c *sqlConnection) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	// total sum is 126 seconds (sum(2*2^x) from 0 to 5)
	re := retrier.NewRetrier(6, 2*time.Second, &coordinatorHTTPCheck{
		c: c,
	})
	return re.RunCtx(ctx, func(ctx2 context.Context) (driver.Rows, retrier.Action, error) {
		dr := newDruidRequest(query, args)
		b, err := json.Marshal(dr)
		if err != nil {
			return nil, retrier.Fail, err
		}

		bodyReader := bytes.NewReader(b)

		context.AfterFunc(ctx, func() {
			tctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			r, err := http.NewRequestWithContext(tctx, http.MethodDelete, c.dsn+"/"+dr.Context.SQLQueryID, http.NoBody)
			if err != nil {
				return
			}

			resp, err := c.client.Do(r)
			if err != nil {
				return
			}
			resp.Body.Close()
		})

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.dsn, bodyReader)
		if err != nil {
			return nil, retrier.Fail, err
		}

		req.Header.Add("Content-Type", "application/json")
		resp, err := c.client.Do(req)
		if err != nil {
			// nolint:errorlint // there's no wrapping
			if v, ok := err.(*url.Error); ok {
				// Don't retry if the error was due to too many redirects.
				if redirectsErrorRe.MatchString(v.Error()) {
					return nil, retrier.Fail, v
				}

				// Don't retry if the error was due to an invalid protocol scheme.
				if schemeErrorRe.MatchString(v.Error()) {
					return nil, retrier.Fail, v
				}

				// Don't retry if the error was due to TLS cert verification failure.
				if notTrustedErrorRe.MatchString(v.Error()) {
					return nil, retrier.Fail, v
				}

				// nolint:errorlint // there's no wrapping
				if _, ok := v.Err.(x509.UnknownAuthorityError); ok {
					return nil, retrier.Fail, v
				}
			}

			return nil, retrier.Retry, err
		}

		switch resp.StatusCode {
		case http.StatusTooManyRequests:
			return nil, retrier.Retry, fmt.Errorf("Too many requests")
		case http.StatusUnauthorized, http.StatusForbidden:
			return nil, retrier.Fail, fmt.Errorf("Unauthorized request")
		}

		dec := json.NewDecoder(resp.Body)

		var obj any
		err = dec.Decode(&obj)
		if err != nil {
			resp.Body.Close()
			return nil, retrier.Fail, err
		}
		switch v := obj.(type) {
		case map[string]any:
			resp.Body.Close()
			a := retrier.Retry
			if v["errorCode"] == "invalidInput" { // hard fail all invalid-syntax errors
				a = retrier.AdditionalCheck
			}
			return nil, a, fmt.Errorf("%v", obj)
		case []any:
			columns := toStringArray(v)
			err = dec.Decode(&obj)
			if err != nil {
				resp.Body.Close()
				return nil, retrier.Fail, err
			}

			types := toStringArray(obj.([]any))

			transformers := make([]func(any) (any, error), len(columns))
			for i, c := range types {
				transformers[i] = identityTransformer
				switch c {
				case "TINYINT":
					transformers[i] = func(v any) (any, error) {
						switch v := v.(type) {
						case float64:
							return int8(v), nil
						default:
							return v, nil
						}
					}
				case "SMALLINT":
					transformers[i] = func(v any) (any, error) {
						switch v := v.(type) {
						case float64:
							return int16(v), nil
						default:
							return v, nil
						}
					}
				case "INTEGER":
					transformers[i] = func(v any) (any, error) {
						switch v := v.(type) {
						case float64:
							return int32(v), nil
						default:
							return v, nil
						}
					}
				case "BIGINT":
					transformers[i] = func(v any) (any, error) {
						switch v := v.(type) {
						case float64:
							return int64(v), nil
						default:
							return v, nil
						}
					}
				case "FLOAT":
					transformers[i] = func(v any) (any, error) {
						switch v := v.(type) {
						case float64:
							return float32(v), nil
						case string:
							return strconv.ParseFloat(v, 32)
						default:
							return v, nil
						}
					}
				case "DOUBLE":
					transformers[i] = func(v any) (any, error) {
						switch v := v.(type) {
						case string:
							return strconv.ParseFloat(v, 64)
						default:
							return v, nil
						}
					}
				case "REAL":
					transformers[i] = func(v any) (any, error) {
						switch v := v.(type) {
						case string:
							return strconv.ParseFloat(v, 64)
						default:
							return v, nil
						}
					}
				case "DECIMAL":
					transformers[i] = func(v any) (any, error) {
						switch v := v.(type) {
						case string:
							return strconv.ParseFloat(v, 64)
						default:
							return v, nil
						}
					}
				case "TIMESTAMP":
					transformers[i] = func(v any) (any, error) {
						switch v := v.(type) {
						case string:
							t, err := time.Parse(time.RFC3339, v)
							if err != nil {
								return nil, err
							}
							return t, nil
						default:
							return v, nil
						}
					}
				case "ARRAY":
					transformers[i] = func(v any) (any, error) {
						var l []any
						err := json.Unmarshal([]byte(v.(string)), &l)
						if err != nil {
							return nil, err
						}
						return l, nil
					}
				case "OTHER":
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
				closer:        resp.Body,
				dec:           dec,
				columns:       columns,
				types:         types,
				transformers:  transformers,
				currentValues: make([]any, len(columns)),
			}
			return druidRows, retrier.Succeed, nil
		default:
			resp.Body.Close()
			return nil, retrier.Fail, fmt.Errorf("unexpected response: %v", obj)
		}
	})
}

type druidRows struct {
	closer        io.ReadCloser
	dec           *json.Decoder
	columns       []string
	types         []string
	transformers  []func(any) (any, error)
	currentValues []any
}

var (
	_ driver.Rows                           = &druidRows{}
	_ driver.RowsColumnTypeDatabaseTypeName = &druidRows{}
)

func (dr *druidRows) Columns() []string {
	return dr.columns
}

func (dr *druidRows) Close() error {
	return dr.closer.Close()
}

func (dr *druidRows) Next(dest []driver.Value) error {
	err := dr.dec.Decode(&dr.currentValues)
	if err != nil {
		return err
	}

	for i, v := range dr.currentValues {
		v, err := dr.transformers[i](v)
		if err != nil {
			return err
		}

		dest[i] = v
	}

	return nil
}

func (dr *druidRows) ColumnTypeScanType(index int) reflect.Type {
	switch dr.types[index] {
	case "BOOLEAN":
		return reflect.TypeOf(true)
	case "TINYINT":
		return reflect.TypeOf(int8(0))
	case "SMALLINT":
		return reflect.TypeOf(int16(0))
	case "INTEGER":
		return reflect.TypeOf(int32(0))
	case "BIGINT":
		return reflect.TypeOf(int64(0))
	case "FLOAT":
		return reflect.TypeOf(float32(0))
	case "DOUBLE":
		return reflect.TypeOf(float64(0))
	case "REAL":
		return reflect.TypeOf(float64(0))
	case "DECIMAL":
		return reflect.TypeOf(float64(0))
	case "CHAR":
		return reflect.TypeOf("")
	case "VARCHAR":
		return reflect.TypeOf("")
	case "TIMESTAMP":
		return reflect.TypeOf(time.Time{})
	case "DATE":
		return reflect.TypeOf(time.Time{})
	case "OTHER":
		return reflect.TypeOf("")
	}
	return nil
}

func (dr *druidRows) ColumnTypeDatabaseTypeName(index int) string {
	return dr.types[index]
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

func newDruidRequest(query string, args []driver.NamedValue) *DruidRequest {
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

func identityTransformer(v any) (any, error) {
	return v, nil
}

func toStringArray(values []any) []string {
	s := make([]string, len(values))
	for i, v := range values {
		vv, _ := v.(string)
		s[i] = vv
	}
	return s
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
