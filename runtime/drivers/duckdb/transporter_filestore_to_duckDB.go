package duckdb

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

func (t *fileStoreToDuckDB) Transfer(ctx context.Context, srcProps, sinkProps map[string]any, opts *drivers.TransferOptions) error {
	sinkCfg, err := parseSinkProperties(sinkProps)
	if err != nil {
		return err
	}

	srcCfg, err := parseFileSourceProperties(srcProps)
	if err != nil {
		return err
	}

	t.logger = t.logger.With(zap.String("source", sinkCfg.Table))

	localPaths, err := t.from.FilePaths(ctx, srcProps)
	if err != nil {
		return err
	}

	if len(localPaths) == 0 {
		return fmt.Errorf("no files to ingest")
	}

	size := fileSize(localPaths)
	if !sizeWithinStorageLimits(t.to, size) {
		return drivers.ErrStorageLimitExceeded
	}
	opts.Progress.Target(size, drivers.ProgressUnitByte)

	var format string
	if srcCfg.Format != "" {
		format = fmt.Sprintf(".%s", srcCfg.Format)
	} else {
		format = fileutil.FullExt(localPaths[0])
	}

	// Ingest data
	from, err := sourceReader(localPaths, format, srcCfg.DuckDB)
	if err != nil {
		return err
	}

	err = t.to.CreateTableAsSelect(ctx, sinkCfg.Table, false, fmt.Sprintf("SELECT * FROM %s", from))
	if err != nil {
		return err
	}
	opts.Progress.Observe(size, drivers.ProgressUnitByte)
	return nil
}
