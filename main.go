package main

// import "C"

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/apache/arrow/go/v11/parquet"
	"github.com/apache/arrow/go/v11/parquet/compress"
	"github.com/apache/arrow/go/v11/parquet/pqarrow"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/bigquery"
	_ "github.com/rilldata/rill/runtime/drivers/bigquery"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	"go.uber.org/zap"
)

func main() {
	ingest()
	h, err := drivers.Open("duckdb", map[string]any{"dsn": "stage_bq.db"}, false, zap.NewNop())
	if err != nil {
		log.Fatal(err)
	}

	olap, _ := h.AsOLAP("")
	err = olap.Exec(context.Background(), &drivers.Statement{Query: "CREATE OR REPLACE TABLE cv AS SELECT * FROM bigquerypq.parquet"})
	if err != nil {
		log.Fatal(err)
	}
}

func ingest() {
	t := time.Now()
	defer func() {
		fmt.Printf("ingestion took %v", time.Since(t).Seconds())
	}()
	fw, err := os.OpenFile("bigquerypq.parquet", os.O_RDWR|os.O_TRUNC|os.O_EXCL, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer fw.Close()

	bq, err := drivers.Open("bigquery", map[string]any{"allow_host_access": true}, false, zap.NewNop())
	if err != nil {
		log.Fatal(err)
	}
	defer bq.Close()

	sql, _ := bq.AsSQLStore()
	r, err := sql.Query(context.Background(), map[string]any{"project_id": "rilldata"}, "SELECT * FROM `bigquery-public-data.covid19_open_data.compatibility_view` LIMIT 10000000")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	rdr, err := bigquery.AsArrowRecordReader(r)
	if err != nil {
		log.Fatal(err)
	}

	pqwriter, err := pqarrow.NewFileWriter(rdr.Schema(), fw, parquet.NewWriterProperties(parquet.WithCompression(compress.Codecs.Snappy)), pqarrow.DefaultWriterProps())
	if err != nil {
		log.Fatal(err)
	}
	defer pqwriter.Close()

	for rdr.Next() {
		if err := pqwriter.WriteBuffered(rdr.Record()); err != nil {
			log.Fatal(err)
		}
	}

}
