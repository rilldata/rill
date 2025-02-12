package sqldriver

import (
	"context"
	"database/sql"
	sqlDriver "database/sql/driver"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/url"
	"reflect"
	"time"

	"github.com/startreedata/pinot-client-go/pinot"
)

type pinotDriver struct{}

func (d *pinotDriver) Open(dsn string) (sqlDriver.Conn, error) {
	broker, _, headers, err := ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	pinotConn, err := pinot.NewWithConfig(&pinot.ClientConfig{
		ExtraHTTPHeader:     headers,
		BrokerList:          []string{broker},
		UseMultistageEngine: true, // We have joins and nested queries which are supported by multistage engine
	})
	if err != nil {
		return nil, err
	}

	return &conn{pinotConn: pinotConn}, nil
}

func init() {
	sql.Register("pinot", &pinotDriver{})
}

type conn struct {
	pinotConn *pinot.Connection
}

func (c *conn) Prepare(query string) (sqlDriver.Stmt, error) {
	return nil, fmt.Errorf("unsupported")
}

func (c *conn) Close() error {
	return nil
}

func (c *conn) Begin() (sqlDriver.Tx, error) {
	return nil, fmt.Errorf("unsupported")
}

func (c *conn) QueryContext(ctx context.Context, query string, args []sqlDriver.NamedValue) (sqlDriver.Rows, error) {
	var resp *pinot.BrokerResponse
	var err error
	if len(args) > 0 {
		var params []interface{}
		for _, arg := range args {
			params = append(params, arg.Value)
		}
		// TODO: cancel the query if ctx is done
		resp, err = c.pinotConn.ExecuteSQLWithParams("", query, params)
	} else {
		resp, err = c.pinotConn.ExecuteSQL("", query)
	}
	if err != nil {
		return nil, err
	}
	if len(resp.Exceptions) > 0 {
		if len(resp.Exceptions) == 1 {
			return nil, fmt.Errorf("query error: %q: %q", resp.Exceptions[0].ErrorCode, resp.Exceptions[0].Message)
		}
		errMsg := "query errors:\n"
		for _, e := range resp.Exceptions {
			errMsg += fmt.Sprintf("\t%q: %q\n", e.ErrorCode, e.Message)
		}
		return nil, errors.New(errMsg)
	}

	cols := colSchema(resp.ResultTable)

	return &rows{results: resp.ResultTable, columns: cols, numRows: resp.ResultTable.GetRowCount(), currIdx: 0}, nil
}

func (c *conn) ExecContext(ctx context.Context, query string, args []sqlDriver.NamedValue) (sqlDriver.Result, error) {
	return nil, fmt.Errorf("unsupported")
}

func (c *conn) Ping(ctx context.Context) error {
	rows, err := c.QueryContext(ctx, "SELECT 1", nil)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

type rows struct {
	results *pinot.ResultTable
	columns []column
	numRows int
	currIdx int
}

func (r *rows) Columns() []string {
	return r.results.DataSchema.ColumnNames
}

func (r *rows) Close() error {
	return nil
}

func (r *rows) Next(dest []sqlDriver.Value) error {
	if r.currIdx >= r.numRows {
		return io.EOF
	}
	for i := range len(r.Columns()) {
		dest[i] = r.goValue(r.currIdx, i, r.results.GetColumnDataType(i))
	}
	r.currIdx++
	return nil
}

func (r *rows) ColumnTypeScanType(index int) reflect.Type {
	return r.columns[index].scanType
}

func (r *rows) ColumnTypeDatabaseTypeName(index int) string {
	return r.columns[index].pinotType
}

type column struct {
	name      string
	pinotType string
	scanType  reflect.Type
}

func colSchema(results *pinot.ResultTable) []column {
	var cols []column
	for i := 0; i < results.GetColumnCount(); i++ {
		cols = append(cols, column{
			name:      results.GetColumnName(i),
			pinotType: results.GetColumnDataType(i),
			scanType:  scanType(results.GetColumnDataType(i)),
		})
	}
	return cols
}

func scanType(pinotType string) reflect.Type {
	switch pinotType {
	case "INT":
		return reflect.TypeOf(int32(0))
	case "LONG":
		return reflect.TypeOf(int64(0))
	case "FLOAT":
		return reflect.TypeOf(float32(0))
	case "DOUBLE":
		return reflect.TypeOf(float64(0))
	case "STRING":
		return reflect.TypeOf("")
	case "BYTES":
		return reflect.TypeOf("")
	case "BIG_DECIMAL":
		return reflect.TypeOf(big.Float{})
	case "TIMESTAMP":
		return reflect.TypeOf(time.Time{})
	case "BOOLEAN":
		return reflect.TypeOf(false)
	default:
		return reflect.TypeOf("")
	}
}

func (r *rows) goValue(rowIdx, coldIdx int, pinotType string) interface{} {
	if r.results.Get(rowIdx, coldIdx) == nil {
		return nil
	}
	switch pinotType {
	case "INT":
		// check if interface is string as it may be NaN
		if reflect.TypeOf(r.results.Get(rowIdx, coldIdx)).String() == "string" {
			return int32(math.NaN())
		}
		return r.results.GetInt(rowIdx, coldIdx)
	case "LONG":
		if reflect.TypeOf(r.results.Get(rowIdx, coldIdx)).String() == "string" {
			return int64(math.NaN())
		}
		return r.results.GetLong(rowIdx, coldIdx)
	case "FLOAT":
		if reflect.TypeOf(r.results.Get(rowIdx, coldIdx)).String() == "string" {
			return float32(math.NaN())
		}
		return r.results.GetFloat(rowIdx, coldIdx)
	case "DOUBLE":
		if reflect.TypeOf(r.results.Get(rowIdx, coldIdx)).String() == "string" {
			return math.NaN()
		}
		return r.results.GetDouble(rowIdx, coldIdx)
	case "STRING":
		return r.results.GetString(rowIdx, coldIdx)
	case "BYTES":
		// return hex string as it is
		return r.results.GetString(rowIdx, coldIdx)
	case "BIG_DECIMAL":
		return r.results.Get(rowIdx, coldIdx)
	case "TIMESTAMP":
		// convert iso8601 formatted string to time.Time
		t, err := time.Parse("2006-01-02 15:04:05.0", r.results.GetString(rowIdx, coldIdx))
		if err != nil {
			return err
		}
		return t
	case "BOOLEAN":
		return r.results.Get(rowIdx, coldIdx).(bool)
	default:
		return reflect.TypeOf("")
	}
}

// ParseDSN parses the DSN string to extract the controller address and basic auth credentials
func ParseDSN(dsn string) (string, string, map[string]string, error) {
	// DSN format: http(s)://username:password@broker:port?controller=http(s)://controller:port
	// validate dsn - it should be a valid URL, may contain basic auth credentials
	u, err := url.Parse(dsn)
	if err != nil {
		return "", "", nil, fmt.Errorf("invalid DSN: %w", err)
	}

	var authHeader map[string]string
	if u.User != nil {
		uname := u.User.Username()
		pwd, passwordSet := u.User.Password()
		if uname == "" || !passwordSet {
			return "", "", nil, fmt.Errorf("DSN should contain valid basic auth credentials")
		}
		// clear user info from URL so that u.String() doesn't include it
		u.User = nil
		authString := fmt.Sprintf("%s:%s", uname, pwd)
		authHeader = map[string]string{
			"Authorization": fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(authString))),
		}
	}

	controllerURL := u.Query().Get("controller")
	if controllerURL == "" {
		return "", "", nil, fmt.Errorf("controller URL not provided, dsn is form http(s)://username:password@broker:port?controller=http(s)://controller:port")
	}

	u.RawQuery = ""
	return u.String(), controllerURL, authHeader, nil
}
