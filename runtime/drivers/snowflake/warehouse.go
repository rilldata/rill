package snowflake

import (
	"context"
	"database/sql"
	sqld "database/sql/driver"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/apache/arrow/go/v15/parquet"
	"github.com/apache/arrow/go/v15/parquet/compress"
	"github.com/apache/arrow/go/v15/parquet/pqarrow"
	"github.com/c2h5oh/datasize"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	sf "github.com/snowflakedb/gosnowflake"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// entire data is buffered in memory before its written to disk so keeping it small reduces memory usage
// but keeping it too small can lead to bad ingestion performance
// 64MB seems to be a good balance
const rowGroupBufferSize = int64(datasize.MB) * 64

// QueryAsFiles implements drivers.SQLStore.
// Fetches query result in arrow batches.
// As an alternative (or in case of memory issues) consider utilizing Snowflake "COPY INTO <location>" feature,
// see https://docs.snowflake.com/en/sql-reference/sql/copy-into-location
func (c *connection) QueryAsFiles(ctx context.Context, props map[string]any) (iter drivers.FileIterator, resErr error) {
	srcProps, err := parseSourceProperties(props)
	if err != nil {
		return nil, err
	}

	var dsn string
	if srcProps.DSN != "" { // get from src properties
		dsn = srcProps.DSN
	} else if c.configProperties.DSN != "" { // get from driver configs
		dsn = c.configProperties.DSN
	} else {
		return nil, fmt.Errorf("the property 'dsn' is required for Snowflake: configure it in the YAML properties or set the 'connector.snowflake.dsn' environment variable")
	}

	parallelFetchLimit := 5
	if c.configProperties.ParallelFetchLimit != 0 {
		parallelFetchLimit = c.configProperties.ParallelFetchLimit
	}

	db, err := sql.Open("snowflake", dsn)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resErr != nil {
			db.Close()
		}
	}()

	ctx = sf.WithArrowAllocator(sf.WithArrowBatches(ctx), memory.DefaultAllocator)

	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resErr != nil {
			conn.Close()
		}
	}()

	var rows sqld.Rows
	err = conn.Raw(func(x interface{}) error {
		rows, err = x.(sqld.QueryerContext).QueryContext(ctx, srcProps.SQL, nil)
		return err
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		if resErr != nil {
			rows.Close()
		}
	}()

	batches, err := rows.(sf.SnowflakeRows).GetArrowBatches()
	if err != nil {
		return nil, err
	}

	if len(batches) == 0 {
		// empty result
		return nil, drivers.ErrNoRows
	}

	for i := range batches {
		batches[i] = batches[i].WithContext(ctx)
	}

	tempDir, err := c.storage.RandomTempDir("snowflake")
	if err != nil {
		return nil, err
	}
	return &fileIterator{
		ctx:                ctx,
		db:                 db,
		conn:               conn,
		rows:               rows,
		batches:            batches,
		parallelFetchLimit: parallelFetchLimit,
		logger:             c.logger,
		tempDir:            tempDir,
	}, nil
}

type fileIterator struct {
	ctx     context.Context
	db      *sql.DB
	conn    *sql.Conn
	rows    sqld.Rows
	batches []*sf.ArrowBatch
	logger  *zap.Logger
	tempDir string
	// Computed while iterating
	totalRecords int64
	downloaded   bool
	// Max number of batches to fetch in parallel
	parallelFetchLimit int
}

// Close implements drivers.FileIterator.
func (f *fileIterator) Close() error {
	if f.rows != nil {
		f.rows.Close()
		f.conn.Close()
		f.db.Close()
	}
	return os.RemoveAll(f.tempDir)
}

// Next implements drivers.FileIterator.
// Query result is written to a single parquet file.
func (f *fileIterator) Next() ([]string, error) {
	if f.downloaded {
		return nil, io.EOF
	}

	// close db resources early
	defer func() {
		f.rows.Close()
		f.conn.Close()
		f.db.Close()
		// mark rows as nil to prevent double close
		f.rows = nil
	}()

	f.logger.Debug("downloading results in parquet file", observability.ZapCtx(f.ctx))

	// create a temp file
	fw, err := os.CreateTemp(f.tempDir, "temp*.parquet")
	if err != nil {
		return nil, err
	}
	defer fw.Close()
	f.downloaded = true

	tf := time.Now()
	defer func() {
		f.logger.Debug("time taken to write arrow records in parquet file", zap.Duration("duration", time.Since(tf)), observability.ZapCtx(f.ctx))
	}()

	firstBatch, err := f.batches[0].Fetch()
	if err != nil {
		return nil, err
	}

	if len(*firstBatch) == 0 {
		// empty result
		return nil, drivers.ErrNoRows
	}

	// common schema
	schema := (*firstBatch)[0].Schema()
	for _, f := range schema.Fields() {
		if f.Type.ID() == arrow.TIME32 || f.Type.ID() == arrow.TIME64 {
			return nil, fmt.Errorf("TIME data type (column %q) is not currently supported, "+
				"consider excluding it or casting it to another data type", f.Name)
		}
	}

	writer, err := pqarrow.NewFileWriter(schema, fw,
		parquet.NewWriterProperties(
			parquet.WithCompression(compress.Codecs.Snappy),
			parquet.WithRootRepetition(parquet.Repetitions.Required),
			// duckdb has issues reading statistics of string type generated with this write
			// column statistics may not be useful if full file need to be ingested so better to disable to save computations
			parquet.WithStats(false),
		),
		pqarrow.NewArrowWriterProperties(pqarrow.WithStoreSchema()))
	if err != nil {
		return nil, err
	}
	defer writer.Close()

	// write arrow records to parquet file
	// the following iteration might be memory intensive
	// since batches are organized as a slice and every batch caches its content
	f.logger.Debug("starting to fetch and process arrow batches",
		zap.Int("batches", len(f.batches)), zap.Int("parallel_fetch_limit", f.parallelFetchLimit), observability.ZapCtx(f.ctx))

	// Fetch batches async
	errGrp, _ := errgroup.WithContext(f.ctx)
	errGrp.SetLimit(f.parallelFetchLimit)
	// mutex to protect file writes
	var mu sync.Mutex
	batchesLeft := len(f.batches)
	start := time.Now()

	for _, batch := range f.batches {
		b := batch
		errGrp.Go(func() error {
			records, err := b.Fetch()
			if err != nil {
				return err
			}
			mu.Lock()
			defer mu.Unlock()

			for _, rec := range *records {
				if writer.RowGroupTotalBytesWritten() >= rowGroupBufferSize {
					writer.NewBufferedRowGroup()
					f.logger.Debug(
						"starting writing to new parquet row group",
						zap.Float64("progress", float64(len(f.batches)-batchesLeft)/float64(len(f.batches))*100),
						zap.Int("total_records", int(f.totalRecords)),
						zap.Duration("elapsed", time.Since(start)), observability.ZapCtx(f.ctx),
					)
				}
				if err := writer.WriteBuffered(rec); err != nil {
					return err
				}
			}
			batchesLeft--
			f.totalRecords += int64(b.GetRowCount())
			return nil
		})
	}

	if err := errGrp.Wait(); err != nil {
		return nil, err
	}

	writer.Close()
	fw.Close()

	fileInfo, err := os.Stat(fw.Name())
	if err != nil {
		return nil, err
	}
	f.logger.Debug("size of file", zap.String("size", datasize.ByteSize(fileInfo.Size()).HumanReadable()), observability.ZapCtx(f.ctx))
	return []string{fw.Name()}, nil
}

func (f *fileIterator) Format() string {
	return ""
}

var _ drivers.FileIterator = &fileIterator{}

type sourceProperties struct {
	SQL string `mapstructure:"sql"`
	DSN string `mapstructure:"dsn"`
}

func parseSourceProperties(props map[string]any) (*sourceProperties, error) {
	conf := &sourceProperties{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}
	if conf.SQL == "" {
		return nil, fmt.Errorf("property 'sql' is mandatory for connector \"snowflake\"")
	}
	return conf, err
}
