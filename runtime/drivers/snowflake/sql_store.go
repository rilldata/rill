package snowflake

import (
	"context"
	"database/sql"
	sqld "database/sql/driver"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/memory"
	"github.com/apache/arrow/go/v14/parquet"
	"github.com/apache/arrow/go/v14/parquet/compress"
	"github.com/apache/arrow/go/v14/parquet/pqarrow"
	"github.com/c2h5oh/datasize"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	sf "github.com/snowflakedb/gosnowflake"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// recommended size is 512MB - 1GB, entire data is buffered in memory before its written to disk
const rowGroupBufferSize = int64(datasize.MB) * 512

// Query implements drivers.SQLStore
func (c *connection) Query(ctx context.Context, props map[string]any) (drivers.RowIterator, error) {
	return nil, drivers.ErrNotImplemented
}

// QueryAsFiles implements drivers.SQLStore.
// Fetches query result in arrow batches.
// As an alternative (or in case of memory issues) consider utilizing Snowflake "COPY INTO <location>" feature,
// see https://docs.snowflake.com/en/sql-reference/sql/copy-into-location
func (c *connection) QueryAsFiles(ctx context.Context, props map[string]any, opt *drivers.QueryOption, p drivers.Progress) (drivers.FileIterator, error) {
	srcProps, err := parseSourceProperties(props)
	if err != nil {
		return nil, err
	}

	var dsn string
	if srcProps.DSN != "" { // get from src properties
		dsn = srcProps.DSN
	} else if url, ok := c.config["dsn"].(string); ok && url != "" { // get from driver configs
		dsn = url
	} else {
		return nil, fmt.Errorf("the property 'dsn' is required for Snowflake. Provide 'dsn' in the YAML properties or pass '--var connector.snowflake.dsn=...' to 'rill start'")
	}

	parallelFetchLimit := 15
	if limit, ok := c.config["parallel_fetch_limit"].(string); ok {
		parallelFetchLimit, err = strconv.Atoi(limit)
		if err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("snowflake", dsn)
	if err != nil {
		return nil, err
	}

	ctx = sf.WithArrowAllocator(sf.WithArrowBatches(ctx), memory.DefaultAllocator)

	conn, err := db.Conn(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	var rows sqld.Rows
	err = conn.Raw(func(x interface{}) error {
		rows, err = x.(sqld.QueryerContext).QueryContext(ctx, srcProps.SQL, nil)
		return err
	})
	if err != nil {
		conn.Close()
		db.Close()
		return nil, err
	}

	batches, err := rows.(sf.SnowflakeRows).GetArrowBatches()
	if err != nil {
		return nil, err
	}

	if len(batches) == 0 {
		// empty result
		return nil, fmt.Errorf("no results found for the query")
	}

	// the number of returned rows is unknown at this point, only the number of batches and output files
	p.Target(1, drivers.ProgressUnitFile)

	return &fileIterator{
		ctx:                ctx,
		db:                 db,
		conn:               conn,
		rows:               rows,
		batches:            batches,
		progress:           p,
		limitInBytes:       opt.TotalLimitInBytes,
		parallelFetchLimit: parallelFetchLimit,
		logger:             c.logger,
	}, nil
}

type fileIterator struct {
	ctx          context.Context
	db           *sql.DB
	conn         *sql.Conn
	rows         sqld.Rows
	batches      []*sf.ArrowBatch
	progress     drivers.Progress
	limitInBytes int64
	logger       *zap.Logger
	// Computed while iterating
	totalRecords int64
	tempFilePath string
	downloaded   bool
	// Max number of batches to fetch in parallel
	parallelFetchLimit int
}

// Close implements drivers.FileIterator.
func (f *fileIterator) Close() error {
	return os.Remove(f.tempFilePath)
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
	}()

	f.logger.Debug("downloading results in parquet file", observability.ZapCtx(f.ctx))

	// create a temp file
	fw, err := os.CreateTemp("", "temp*.parquet")
	if err != nil {
		return nil, err
	}
	defer fw.Close()
	f.tempFilePath = fw.Name()
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
		return nil, fmt.Errorf("no results found for the query")
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
		zap.Int("batches", len(f.batches)), zap.Int("parallel_fetch_limit", f.parallelFetchLimit))

	// Fetch batches async
	fetchGrp, ctx := errgroup.WithContext(f.ctx)
	fetchGrp.SetLimit(f.parallelFetchLimit)
	fetchResultChan := make(chan fetchResult)

	// Write batches into a file async
	writeGrp, _ := errgroup.WithContext(f.ctx)
	writeGrp.Go(func() error {
		batchesLeft := len(f.batches)
		for {
			select {
			case result, ok := <-fetchResultChan:
				if !ok {
					return nil
				}
				batch := result.batch
				writeStart := time.Now()
				for _, rec := range *result.records {
					if writer.RowGroupTotalBytesWritten() >= rowGroupBufferSize {
						writer.NewBufferedRowGroup()
					}
					if err := writer.WriteBuffered(rec); err != nil {
						return err
					}
					fileInfo, err := os.Stat(fw.Name())
					if err == nil { // ignore error
						if fileInfo.Size() > f.limitInBytes {
							return drivers.ErrStorageLimitExceeded
						}
					}
				}
				batchesLeft--
				f.logger.Debug(
					"wrote an arrow batch to a parquet file",
					zap.Float64("progress", float64(len(f.batches)-batchesLeft)/float64(len(f.batches))*100),
					zap.Int("row_count", batch.GetRowCount()),
					zap.Duration("write_duration", time.Since(writeStart)),
				)
				f.totalRecords += int64(result.batch.GetRowCount())
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})

	for _, batch := range f.batches {
		b := batch
		fetchGrp.Go(func() error {
			fetchStart := time.Now()
			records, err := b.Fetch()
			if err != nil {
				return err
			}
			fetchResultChan <- fetchResult{records: records, batch: b}
			f.logger.Debug(
				"fetched an arrow batch",
				zap.Duration("duration", time.Since(fetchStart)),
				zap.Int("row_count", b.GetRowCount()),
			)
			return nil
		})
	}

	err = fetchGrp.Wait()
	close(fetchResultChan)

	if err != nil {
		return nil, err
	}

	if err := writeGrp.Wait(); err != nil {
		return nil, err
	}

	writer.Close()
	fw.Close()

	fileInfo, err := os.Stat(fw.Name())
	if err != nil {
		return nil, err
	}
	f.progress.Observe(1, drivers.ProgressUnitFile)
	f.logger.Debug("size of file", zap.String("size", datasize.ByteSize(fileInfo.Size()).HumanReadable()), observability.ZapCtx(f.ctx))
	return []string{fw.Name()}, nil
}

// Size implements drivers.FileIterator.
func (f *fileIterator) Size(unit drivers.ProgressUnit) (int64, bool) {
	switch unit {
	case drivers.ProgressUnitFile:
		return 1, true
	// the number of records is unknown until the end of iteration
	case drivers.ProgressUnitRecord:
		return f.totalRecords, true
	default:
		return 0, false
	}
}

func (f *fileIterator) Format() string {
	return ""
}

var _ drivers.FileIterator = &fileIterator{}

type fetchResult struct {
	records *[]arrow.Record
	batch   *sf.ArrowBatch
}

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
