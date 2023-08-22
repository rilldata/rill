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
	"strings"
	"time"
	"unsafe"

	"github.com/apache/arrow/go/v11/arrow/cdata"
	"github.com/marcboeker/go-duckdb"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/bigquery"
	_ "github.com/rilldata/rill/runtime/drivers/bigquery"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	"go.uber.org/zap"
)

// func main() {
// 	ingest()
// 	h, err := drivers.Open("duckdb", map[string]any{"dsn": "stage_bq.db"}, false, zap.NewNop())
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	olap, _ := h.AsOLAP("")
// 	err = olap.Exec(context.Background(), &drivers.Statement{Query: "CREATE OR REPLACE TABLE cv AS SELECT * FROM bigquerypq.parquet"})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

func main() {
	logger := log.Default()
	t := time.Now()
	defer func() {
		logger.Printf("ingestion took %v", time.Since(t).Seconds())
	}()

	h, err := drivers.Open("duckdb", map[string]any{"dsn": "stage_bq.db"}, false, zap.NewNop())
	if err != nil {
		log.Fatal(err)
	}

	// fw, err := os.OpenFile("bigquerypq.parquet", os.O_RDWR|os.O_TRUNC|os.O_EXCL, 0600)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer fw.Close()

	bq, err := drivers.Open("bigquery", map[string]any{"allow_host_access": true}, false, zap.NewNop())
	if err != nil {
		log.Fatal(err)
	}
	defer bq.Close()

	tw := time.Now()
	s, _ := bq.AsSQLStore()
	r, err := s.Query(context.Background(), map[string]any{"project_id": "rilldata"}, "SELECT * FROM `bigquery-public-data.covid19_open_data.compatibility_view` LIMIT 10000000")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	logger.Printf("query took seconds %v\n", time.Since(tw).Seconds())

	rdr, err := bigquery.AsArrowRecordReader(r)
	if err != nil {
		log.Fatal(err)
	}
	defer rdr.Release()

	olap, _ := h.AsOLAP("")
	err = olap.WithConnection(context.Background(), 1, func(ctx, ensuredCtx context.Context, conn *sql.Conn) error {
		return rawConn(conn, func(conn driver.Conn) error {
			a, err := duckdb.NewArrowQueryFromConn(conn)
			if err != nil {
				return err
			}

			var qry, scan, memory time.Duration
			defer func() {
				logger.Printf("scan took seconds %v\n", scan.Seconds())
				logger.Printf("insert took seconds %v\n", qry.Seconds())
				logger.Printf("memory took seconds %v\n", qry.Seconds())
			}()
			views := make([]string, 0)
			i := 0
			for rdr.Next() {
				// fmt.Println("appending a record")
				rec := rdr.Record()
				tm := time.Now()
				var arrowArray = C.calloc(1, C.sizeof_struct_ArrowArray)
				defer C.free(arrowArray)
				pArrowArray := (*cdata.CArrowArray)(unsafe.Pointer(arrowArray))

				var arrowSchema = C.calloc(1, C.sizeof_struct_ArrowSchema)
				defer C.free(arrowSchema)
				pArrowSchema := (*cdata.CArrowSchema)(unsafe.Pointer(arrowSchema))

				cdata.ExportArrowRecordBatch(rec, pArrowArray, pArrowSchema)

				var res = C.calloc(1, C.sizeof_struct_ArrowArrayStream)
				defer C.free(res)
				pres := (*cdata.CArrowArrayStream)(unsafe.Pointer(res))
				memory += time.Since(tm)

				ts := time.Now()
				view := fmt.Sprintf("view%v", i)
				views = append(views, view)
				i++
				err = a.ScanArrowContext(ctx, view, pArrowSchema, pArrowArray, pres)
				if err != nil {
					return err
				}
				scan += time.Since(ts)
			}

			tq := time.Now()
			_, err = a.QueryContext(context.Background(), fmt.Sprintf("CREATE OR REPLACE TABLE t AS %s", query(views)))
			qry += time.Since(tq)
			if err != nil {
				return err
			}

			return nil
		})
	})
	if err != nil {
		log.Fatal(err)
	}

	tx := time.Now()
	row, err := olap.Execute(context.Background(), &drivers.Statement{Query: "select count(*) from t"})
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	row.Next()

	var count int
	err = row.Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(count)
	defer func() {
		logger.Printf("count query took %v", time.Since(tx).Seconds())
	}()
}

// rawConn is similar to *sql.Conn.Raw, but additionally unwraps otelsql (which we use for instrumentation).
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

func query(views []string) string {
	for i, v := range views {
		views[i] = "SELECT * FROM " + v
	}
	return strings.Join(views, " UNION ALL ")
}
