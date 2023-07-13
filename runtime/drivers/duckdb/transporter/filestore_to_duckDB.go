package transporter

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"go.uber.org/zap"
)

type fileStoreToDuckDB struct {
	to     drivers.OLAPStore
	from   drivers.FileStore
	logger *zap.Logger
}

func NewFileStoreToDuckDB(from drivers.FileStore, to drivers.OLAPStore, logger *zap.Logger) drivers.Transporter {
	return &fileStoreToDuckDB{
		to:     to,
		from:   from,
		logger: logger,
	}
}

var _ drivers.Transporter = &fileStoreToDuckDB{}

func (t *fileStoreToDuckDB) Transfer(ctx context.Context, source drivers.Source, sink drivers.Sink, opts *drivers.TransferOpts, p drivers.Progress) error {
	src, ok := source.FileSource()
	if !ok {
		return fmt.Errorf("type of source should `drivers.FilesSource`")
	}
	fSink, ok := sink.DatabaseSink()
	if !ok {
		return fmt.Errorf("type of source should `drivers.DatabaseSink`")
	}

	localPaths, err := t.from.FilePaths(ctx, src)
	if err != nil {
		return err
	}

	if len(localPaths) == 0 {
		return fmt.Errorf("no files to ingest")
	}

	size := fileSize(localPaths)
	if size > opts.LimitInBytes {
		return drivers.ErrIngestionLimitExceeded
	}
	p.Target(size, drivers.ProgressUnitByte)

	var format string
	if val, ok := src.Properties["format"].(string); ok {
		format = fmt.Sprintf(".%s", val)
	} else {
		format = fileutil.FullExt(localPaths[0])
	}

	var ingestionProps map[string]any
	if duckDBProps, ok := src.Properties["duckdb"].(map[string]any); ok {
		ingestionProps = duckDBProps
	} else {
		ingestionProps = map[string]any{}
	}

	// Ingest data
	from, err := sourceReader(localPaths, format, ingestionProps)
	if err != nil {
		return err
	}

	qry := fmt.Sprintf("CREATE OR REPLACE TABLE %q AS (SELECT * FROM %s)", fSink.Table, from)
	err = t.to.Exec(ctx, &drivers.Statement{Query: qry, Priority: 1})
	if err != nil {
		return err
	}
	p.Observe(size, drivers.ProgressUnitByte)
	return nil
}
