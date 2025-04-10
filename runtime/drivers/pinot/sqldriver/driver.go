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
	"net"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/rilldata/rill/runtime/pkg/retrier"
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
	re := retrier.NewRetrier(6, 2*time.Second, nil)
	return re.RunCtx(ctx, func(ctx context.Context) (sqlDriver.Rows, retrier.Action, error) {
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
			if isRetryableHTTPError(err) {
				return nil, retrier.Retry, err
			}
			return nil, retrier.Fail, err
		}
		if len(resp.Exceptions) > 0 {
			errMsg := "query errors:"
			for _, e := range resp.Exceptions {
				errMsg += fmt.Sprintf("\t%d: %q\n", e.ErrorCode, e.Message)
			}
			err := errors.New(errMsg)
			for _, e := range resp.Exceptions {
				if isRetryablePinotErrorCode(e.ErrorCode) {
					return nil, retrier.Retry, err
				}
			}
			return nil, retrier.Fail, err
		}

		cols := colSchema(resp.ResultTable)

		return &rows{results: resp.ResultTable, columns: cols, numRows: resp.ResultTable.GetRowCount(), currIdx: 0}, retrier.Succeed, nil
	})
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

func isRetryableHTTPError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return true
		}
	}
	errStr := err.Error()
	re := regexp.MustCompile(`Pinot: (\d{3})`)
	matches := re.FindStringSubmatch(errStr)
	if len(matches) == 2 {
		if code, convErr := strconv.Atoi(matches[1]); convErr == nil {
			return isRetryableHTTPCode(code)
		}
	}
	return false
}

func isRetryableHTTPCode(code int) bool {
	switch code {
	case 408: // Request Timeout — client didn't produce a request in time
	case 429: // Too Many Requests — server is rate-limiting, often includes Retry-After
	case 502: // Bad Gateway — server got invalid response from upstream
	case 503: // Service Unavailable — server is overloaded or down for maintenance
	case 504: // Gateway Timeout — server acting as a gateway timed out waiting for upstream
	default:
		return false
	}
	return true
}

func isRetryablePinotErrorCode(code int) bool {
	// Pinot code are from https://github.com/apache/pinot/blob/master/pinot-spi/src/main/java/org/apache/pinot/spi/exception/QueryErrorCode.java
	switch code {
	case 210: // SERVER_SHUTTING_DOWN
	case 211: // SERVER_OUT_OF_CAPACITY
	case 240: // QUERY_SCHEDULING_TIMEOUT
	case 245: // SERVER_RESOURCE_LIMIT_EXCEEDED
	case 250: // EXECUTION_TIMEOUT
	case 400: // BROKER_TIMEOUT
	case 427: // SERVER_NOT_RESPONDING
	case 429: // TOO_MANY_REQUESTS
	default:
		return false
	}
	return true
}
