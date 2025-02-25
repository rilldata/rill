package duckdb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

type warehouseToDuckDB struct {
	to     drivers.OLAPStore
	from   drivers.Warehouse
	logger *zap.Logger
}

var _ drivers.Transporter = &warehouseToDuckDB{}

func NewWarehouseToDuckDB(from drivers.Warehouse, to drivers.OLAPStore, logger *zap.Logger) drivers.Transporter {
	return &warehouseToDuckDB{
		from:   from,
		to:     to,
		logger: logger,
	}
}

func (w *warehouseToDuckDB) Transfer(ctx context.Context, srcProps, sinkProps map[string]any, opts *drivers.TransferOptions) (transferErr error) {
	sinkCfg, err := parseSinkProperties(sinkProps)
	if err != nil {
		return err
	}

	w.logger = w.logger.With(zap.String("source", sinkCfg.Table))

	iter, err := w.from.QueryAsFiles(ctx, srcProps)
	if err != nil {
		return err
	}
	defer iter.Close()

	start := time.Now()
	w.logger.Debug("started transfer from local file to duckdb", zap.String("sink_table", sinkCfg.Table), observability.ZapCtx(ctx))
	defer func() {
		w.logger.Debug("transfer finished",
			zap.Duration("duration", time.Since(start)),
			zap.Bool("success", transferErr == nil),
			observability.ZapCtx(ctx))
	}()
	create := true
	// TODO :: iteration over fileiterator is similar(apart from no schema changes possible here)
	// to consuming fileIterator in objectStore_to_duckDB
	// both can be refactored to follow same path
	for {
		files, err := iter.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		format := fileutil.FullExt(files[0])
		if iter.Format() != "" {
			format += "." + iter.Format()
		}

		from, err := sourceReader(files, format, make(map[string]any))
		if err != nil {
			return err
		}

		if create {
			_, err = w.to.CreateTableAsSelect(ctx, sinkCfg.Table, fmt.Sprintf("SELECT * FROM %s", from), &drivers.CreateTableOptions{})
			create = false
		} else {
			insertOpts := &drivers.InsertTableOptions{
				ByName:   false,
				InPlace:  true,
				Strategy: drivers.IncrementalStrategyAppend,
			}
			_, err = w.to.InsertTableAsSelect(ctx, sinkCfg.Table, fmt.Sprintf("SELECT * FROM %s", from), insertOpts)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
