package transporter

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/apache/arrow/go/v13/arrow/array"
	"github.com/apache/arrow/go/v13/parquet"
	"github.com/apache/arrow/go/v13/parquet/compress"
	"github.com/apache/arrow/go/v13/parquet/pqarrow"
	"github.com/c2h5oh/datasize"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

type sqlStoreToDuckDB struct {
	to     drivers.OLAPStore
	from   drivers.SQLStore
	logger *zap.Logger
}

var _ drivers.Transporter = &sqlStoreToDuckDB{}

func NewSQLStoreToDuckDB(from drivers.SQLStore, to drivers.OLAPStore, logger *zap.Logger) drivers.Transporter {
	return &sqlStoreToDuckDB{
		to:     to,
		from:   from,
		logger: logger,
	}
}

func (s *sqlStoreToDuckDB) Transfer(ctx context.Context, source drivers.Source, sink drivers.Sink, opts *drivers.TransferOpts, p drivers.Progress) (transferErr error) {
	src, ok := source.DatabaseSource()
	if !ok {
		return fmt.Errorf("type of source should `drivers.DatabaseSource`")
	}
	dbSink, ok := sink.DatabaseSink()
	if !ok {
		return fmt.Errorf("type of source should `drivers.DatabaseSink`")
	}

	iter, err := s.from.Query(ctx, src.Props, src.SQL)
	if err != nil {
		return err
	}
	defer iter.Close()

	arrReader, err := iter.AsArrowRecordReader()
	if err != nil {
		// TODO :: differentiate b/w not implemented error and other errors and can add support for row by row ingestion in future
		return err
	}
	defer arrReader.Release()

	start := time.Now()
	s.logger.Info("started transfer from SQL store to duckdb", zap.String("sink_table", dbSink.Table), observability.ZapCtx(ctx))
	defer func() {
		s.logger.Info("transfer finished",
			zap.Duration("duration", time.Since(start)),
			zap.Bool("success", transferErr == nil),
			observability.ZapCtx(ctx))
	}()
	return s.downloadAsParquet(ctx, arrReader, dbSink.Table, p)
}

func (s *sqlStoreToDuckDB) downloadAsParquet(ctx context.Context, rdr array.RecordReader, sinkTable string, p drivers.Progress) error {
	// create a temp directory
	tf := time.Now()
	temp, err := os.MkdirTemp(os.TempDir(), "pq_ingestion")
	if err != nil {
		return err
	}
	defer func() {
		os.RemoveAll(temp)
	}()

	// create a temp file
	fw, err := fileutil.OpenTempFileInDir(temp, "temp.parquet")
	if err != nil {
		return err
	}
	defer fw.Close()

	writer, err := pqarrow.NewFileWriter(rdr.Schema(), fw,
		parquet.NewWriterProperties(
			parquet.WithCompression(compress.Codecs.Snappy),
			parquet.WithRootRepetition(parquet.Repetitions.Required),
			// duckdb has issues reading statistics of string type generated with this write
			// column statistics are not useful if full file need to be ingested so better to disable to save computations
			parquet.WithStats(false),
		),
		pqarrow.NewArrowWriterProperties(pqarrow.WithStoreSchema()))
	if err != nil {
		return err
	}
	defer writer.Close()

	// write arrow records to parquet file
	for rdr.Next() {
		fw.Stat()
		rec := rdr.Record()
		p.Observe(rec.NumRows(), drivers.ProgressUnitRecord)
		if err := writer.WriteBuffered(rec); err != nil {
			return err
		}
	}
	writer.Close()
	fw.Close()
	s.logger.Info("time taken to write parquet file", zap.Duration("duration", time.Since(tf)))

	fileInfo, err := os.Stat(fw.Name())
	if err != nil {
		return err
	}
	s.logger.Info("size of file", zap.String("size", datasize.ByteSize(fileInfo.Size()).HumanReadable()))

	// generate source statement
	from, err := sourceReader([]string{fw.Name()}, "parquet", make(map[string]any))
	if err != nil {
		return err
	}

	ti := time.Now()
	defer func() {
		s.logger.Info("time taken to ingest parquet file", zap.Duration("duration", time.Since(ti)))
	}()
	query := fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM %s);", sinkTable, from)
	return s.to.Exec(ctx, &drivers.Statement{Query: query, Priority: 1})
}
