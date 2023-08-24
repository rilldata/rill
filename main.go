package main

/*
#include <stdlib.h>
#include <arrow.h>
*/
import "C"

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log"
	"time"
	"unsafe"

	"github.com/apache/arrow/go/v13/arrow/cdata"
	"github.com/marcboeker/go-duckdb"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/bigquery"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	"go.uber.org/zap"
)

var qry string = "SELECT * FROM `bigquery-public-data.samples.gsod` LIMIT 50000"

func rawConn(conn *sql.Conn, f func(driver.Conn) error) error {
	return conn.Raw(func(raw any) error {
		// For details, see: https://github.com/XSAM/otelsql/issues/98
		if c, ok := raw.(interface{ Raw() driver.Conn }); ok {
			raw = c.Raw()
		}

		// This is currently guaranteed, but adding check to be safe
		driverConn, ok := raw.(driver.Conn)
		if !ok {
			return fmt.Errorf("internal: did not obtain a driver.Conn")
		}

		return f(driverConn)
	})
}
func main() {
	t := time.Now()
	defer func() {
		log.Printf("ingest took %v seconds", time.Since(t).Seconds())
	}()

	db, err := drivers.Open("duckdb", map[string]any{"dsn": "test1.db"}, false, zap.NewNop())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	olap, _ := db.AsOLAP("")
	olap.WithConnection(context.Background(), 1, func(wrappedCtx, ensuredCtx context.Context, conn *sql.Conn) error {
		getStream(conn)
		return nil
	})
}

func getStream(conn *sql.Conn) {
	bq, err := drivers.Open("bigquery", map[string]any{"allow_host_access": true}, false, zap.NewNop())
	if err != nil {
		log.Fatal(err)
	}
	defer bq.Close()

	tw := time.Now()
	s, _ := bq.AsSQLStore()
	r, err := s.Query(context.Background(), map[string]any{"project_id": "rilldata"}, qry)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	log.Printf("query took seconds %v\n", time.Since(tw).Seconds())

	rdr, err := bigquery.AsArrowRecordReader(r)
	if err != nil {
		log.Fatal(err)
	}
	defer rdr.Release()

	err = conn.Raw(func(driverConn any) error {
		if c, ok := driverConn.(interface{ Raw() driver.Conn }); ok {
			driverConn = c.Raw()
		}

		// This is currently guaranteed, but adding check to be safe
		dConn, ok := driverConn.(driver.Conn)
		if !ok {
			return fmt.Errorf("internal: did not obtain a driver.Conn")
		}

		a, err := duckdb.NewArrowQueryFromConn(dConn)
		if err != nil {
			return err
		}

		ts := time.Now()
		var res = C.calloc(1, C.sizeof_struct_ArrowArrayStream)
		defer C.free(res)
		pres := (*cdata.CArrowArrayStream)(unsafe.Pointer(res))

		cdata.ExportRecordReader(rdr, pres)
		err = a.ScanContext(context.Background(), "view", pres)
		if err != nil {
			return err
		}
		log.Printf("scan took %v seconds", time.Since(ts))

		tq := time.Now()
		_, err = a.QueryContext(context.Background(), "CREATE OR REPLACE TABLE t AS (SELECT * FROM view)")
		log.Printf("insert took %v seconds", time.Since(tq))
		return err
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("getting number of rows")
	rows := conn.QueryRowContext(context.Background(), "select count(*) from t")
	var count int
	err = rows.Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("hello world my rows are " + fmt.Sprint(count))
}
