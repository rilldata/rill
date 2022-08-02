package duckdb

/*
#include <duckdb.h>
*/
import "C"

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type conn struct {
	db     *C.duckdb_database
	con    *C.duckdb_connection
	closed bool
	tx     bool
}

func (c *conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	if c.closed {
		panic("database/sql/driver: misuse of duckdb driver: Exec after Close")
	}
	queryStr, err := c.interpolateParams(query, args)
	if err != nil {
		return nil, err
	}
	res, err := c.exec(queryStr)
	if err != nil {
		return nil, err
	}
	defer C.duckdb_destroy_result(&res)

	ra := int64(C.duckdb_value_int64(&res, 0, 0))

	return &result{ra}, nil
}

func (c *conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	return c.query(query, args)
}

func (c *conn) Prepare(cmd string) (driver.Stmt, error) {
	if c.closed {
		panic("database/sql/driver: misuse of duckdb driver: Prepare after Close")
	}
	cmdstr := C.CString(cmd)
	defer C.free(unsafe.Pointer(cmdstr))

	var s C.duckdb_prepared_statement
	if state := C.duckdb_prepare(*c.con, cmdstr, &s); state == C.DuckDBError {
		dbErr := C.GoString(C.duckdb_prepare_error(s))
		C.duckdb_destroy_prepare(&s)

		return nil, errors.New(dbErr)
	}

	return &stmt{c: c, stmt: &s}, nil
}

func (c *conn) Begin() (driver.Tx, error) {
	if c.tx {
		panic("database/sql/driver: misuse of duckdb driver: multiple Tx")
	}

	if _, err := c.exec("BEGIN TRANSACTION"); err != nil {
		return nil, err
	}

	c.tx = true
	return &tx{c}, nil
}

func (c *conn) Close() error {
	if c.closed {
		panic("database/sql/driver: misuse of duckdb driver: Close of already closed connection")
	}
	c.closed = true

	C.duckdb_disconnect(c.con)
	C.duckdb_close(c.db)
	c.db = nil

	return nil
}

func (c *conn) query(query string, args []driver.Value) (driver.Rows, error) {
	queryStr, err := c.interpolateParams(query, args)
	if err != nil {
		return nil, err
	}

	res, err := c.exec(queryStr)
	if err != nil {
		return nil, err
	}

	return newRows(res), nil
}

func (c *conn) exec(cmd string) (C.duckdb_result, error) {
	cmdstr := C.CString(cmd)
	defer C.free(unsafe.Pointer(cmdstr))

	var res C.duckdb_result

	if err := C.duckdb_query(*c.con, cmdstr, &res); err == C.DuckDBError {
		dbErr := C.duckdb_result_error(&res)
		return res, errors.New(C.GoString(dbErr))
	}

	return res, nil
}

// interpolateParams is taken from
// https://github.com/go-sql-driver/mysql
func (c *conn) interpolateParams(query string, args []driver.Value) (string, error) {
	paramCount := strings.Count(query, "?")
	if paramCount != len(args) {
		return "", fmt.Errorf("invalid number of parameters. expected %d, got %d", paramCount, len(args))
	}

	buf := []byte{}
	argPos := 0

	for i := 0; i < len(query); i++ {
		q := strings.IndexByte(query[i:], '?')
		if q == -1 {
			buf = append(buf, query[i:]...)
			break
		}
		buf = append(buf, query[i:i+q]...)
		i += q

		arg := args[argPos]
		argPos++

		if arg == nil {
			buf = append(buf, "NULL"...)
			continue
		}

		switch v := arg.(type) {
		case int8:
			buf = strconv.AppendInt(buf, int64(v), 10)
		case int16:
			buf = strconv.AppendInt(buf, int64(v), 10)
		case int32:
			buf = strconv.AppendInt(buf, int64(v), 10)
		case int64:
			buf = strconv.AppendInt(buf, int64(v), 10)
		case float64:
			buf = strconv.AppendFloat(buf, v, 'g', -1, 64)
		case bool:
			if v {
				buf = append(buf, '1')
			} else {
				buf = append(buf, '0')
			}
		case time.Time:
			buf = strconv.AppendInt(buf, v.Unix(), 10)
		case string:
			buf = append(buf, '\'')
			buf = append(buf, escapeValue(v)...)
			buf = append(buf, '\'')
		default:
			return "", fmt.Errorf("unknown parameter type %s", v)
		}
	}

	if argPos != len(args) {
		return "", driver.ErrSkip
	}

	return string(buf), nil
}

func escapeValue(v string) []byte {
	buf := bytes.NewBuffer(nil)

	for i := 0; i < len(v); i++ {
		c := v[i]
		switch c {
		case '\x00':
			buf.WriteString("\\\\0")
		case '\n':
			buf.WriteString("\\\\n")
		case '\r':
			buf.WriteString("\\\\r")
		case '\x1a':
			buf.WriteString("\\\\Z")
		case '\'':
			buf.WriteString("\\\\'")
		case '"':
			buf.WriteString("\\\"")
		case '\\':
			buf.WriteString("\\\\")
		default:
			buf.WriteByte(c)
		}
	}

	return buf.Bytes()
}
