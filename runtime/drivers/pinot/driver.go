package pinot

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strings"
	"time"

	sqlDriver "database/sql/driver"
	"github.com/startreedata/pinot-client-go/pinot"
)

type pinotDriver struct{}

func (d *pinotDriver) Open(dsn string) (sqlDriver.Conn, error) {
	// TODO dsn parsing to connect to actual server
	// use pinot.NewWithConfig for extra configs
	connection, err := pinot.NewFromController("localhost:9000")
	if err != nil {
		return nil, err
	}
	// TODO check if it needs to be configurable
	connection.UseMultistageEngine(true)
	return &conn{pinotConn: connection}, nil
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
	if len(args) > 0 {
		q, err := completeQuery(query, args)
		if err != nil {
			return nil, err
		}
		query = q
	}
	// TODO: cancel the query if ctx is done
	resp, err := c.pinotConn.ExecuteSQL("", query)
	if err != nil {
		return nil, err
	}
	if resp.Exceptions != nil && len(resp.Exceptions) > 0 {
		errMsg := "query errors:\n"
		for _, e := range resp.Exceptions {
			errMsg += fmt.Sprintf("\tcode: %q message: %q\n", e.ErrorCode, e.Message)
		}
		return nil, fmt.Errorf(errMsg)
	}

	cols, err := colSchema(resp.ResultTable)
	if err != nil {
		return nil, err
	}

	return &rows{results: resp.ResultTable, columns: cols, numRows: resp.ResultTable.GetRowCount(), currIdx: 0}, nil
}

func (c *conn) ExecContext(ctx context.Context, query string, args []sqlDriver.NamedValue) (sqlDriver.Result, error) {
	return nil, fmt.Errorf("unsupported")
}

func (c *conn) Ping(ctx context.Context) error {
	rows, err := c.QueryContext(ctx, "SELECT 1", nil)
	defer rows.Close()
	return err
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
	row := r.results.Rows[r.currIdx]
	r.currIdx++
	for i, v := range row {
		dest[i] = v
	}
	return nil
}
func (r *rows) ColumnTypeScanType(index int) reflect.Type {
	return r.columns[index].goType
}

func (r *rows) ColumnTypeDatabaseTypeName(index int) string {
	return r.columns[index].pinotType
}

type column struct {
	name      string
	pinotType string
	goType    reflect.Type
}

func colSchema(results *pinot.ResultTable) ([]column, error) {
	var cols []column
	for i := 0; i < results.GetColumnCount(); i++ {
		cols = append(cols, column{
			name:      results.GetColumnName(i),
			pinotType: results.GetColumnDataType(i),
			goType:    pinotToGoType(results.GetColumnDataType(i)),
		})
	}
	return cols, nil
}

func pinotToGoType(pinotType string) reflect.Type {
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
		return reflect.TypeOf([]byte{})
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

func completeQuery(query string, args []sqlDriver.NamedValue) (string, error) {
	parts := strings.Split(query, "?")
	if len(parts)-1 != len(args) {
		return "", fmt.Errorf("mismatch in the number of placeholders and arguments")
	}

	var sb strings.Builder
	for i, part := range parts {
		sb.WriteString(part)
		if i < len(args) {
			argStr, err := formatArg(args[i].Value)
			if err != nil {
				return "", err
			}
			sb.WriteString(argStr)
		}
	}

	return sb.String(), nil
}

// TODO check if it supports all types
func formatArg(value sqlDriver.Value) (string, error) {
	switch v := value.(type) {
	case string, *big.Int, *big.Float:
		// For pinot types - STRING, BIG_DECIMAL and BYTES - enclose in single quotes
		return fmt.Sprintf("'%v'", v), nil
	case []byte:
		// For pinot type - BYTES - convert to Hex string and enclose in single quotes
		hexString := fmt.Sprintf("%x", v)
		return fmt.Sprintf("'%s'", hexString), nil
	case time.Time:
		// For pinot type - TIMESTAMP - convert to ISO8601 format and enclose in single quotes
		return fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05.000Z")), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		// For types - INT, LONG, FLOAT, DOUBLE and BOOLEAN use as-is
		return fmt.Sprintf("%v", v), nil
	default:
		// Throw error for unsupported types
		return "", fmt.Errorf("unsupported type: %T", v)
	}
}
