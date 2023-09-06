package transporter

import (
	"context"
	"fmt"
	"time"

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

	iter, err := s.from.QueryAsFiles(ctx, src.Props, src.SQL, &drivers.QueryOption{TotalLimitInBytes: opts.LimitInBytes}, p)
	if err != nil {
		return err
	}
	defer iter.Close()

	start := time.Now()
	s.logger.Info("started transfer from local file to duckdb", zap.String("sink_table", dbSink.Table), observability.ZapCtx(ctx))
	defer func() {
		s.logger.Info("transfer finished",
			zap.Duration("duration", time.Since(start)),
			zap.Bool("success", transferErr == nil),
			observability.ZapCtx(ctx))
	}()
	create := true
	// TODO :: iteration over fileiterator is similar(apart from no schema changes possible here)
	// to consuming fileIterator in objectStore_to_duckDB
	// both can be refactored to follow same path
	for iter.HasNext() {
		files, err := iter.NextBatch(opts.IteratorBatch)
		if err != nil {
			return err
		}

		format := fileutil.FullExt(files[0])
		from, err := sourceReader(files, format, make(map[string]any), false)
		if err != nil {
			return err
		}

		var query string
		if create {
			query = fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM %s);", safeName(dbSink.Table), from)
			create = false
		} else {
			query = fmt.Sprintf("INSERT INTO %s (SELECT * FROM %s);", safeName(dbSink.Table), from)
		}

		if err := s.to.Exec(ctx, &drivers.Statement{Query: query, Priority: 1}); err != nil {
			return err
		}
	}
	return nil
}
