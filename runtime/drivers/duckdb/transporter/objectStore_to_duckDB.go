package transporter

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

type objectStoreToDuckDB struct {
	to     drivers.OLAPStore
	from   drivers.ObjectStore
	logger *zap.Logger
}

var _ drivers.Transporter = &objectStoreToDuckDB{}

func NewObjectStoreToDuckDB(from drivers.ObjectStore, to drivers.OLAPStore, logger *zap.Logger) drivers.Transporter {
	return &objectStoreToDuckDB{
		to:     to,
		from:   from,
		logger: logger,
	}
}

func (t *objectStoreToDuckDB) Transfer(ctx context.Context, source drivers.Source, sink drivers.Sink, opts *drivers.TransferOpts, p drivers.Progress) error {
	src, ok := source.BucketSource()
	if !ok {
		return fmt.Errorf("type of source should `drivers.BucketSource`")
	}
	dbSink, ok := sink.DatabaseSink()
	if !ok {
		return fmt.Errorf("type of source should `drivers.DatabaseSink`")
	}

	iterator, err := t.from.DownloadFiles(ctx, src)
	if err != nil {
		return err
	}
	defer iterator.Close()

	size, _ := iterator.Size(drivers.ProgressUnitByte)
	if size > opts.LimitInBytes {
		return drivers.ErrIngestionLimitExceeded
	}

	p.Target(size, drivers.ProgressUnitByte)
	var format string
	val, formatDefined := src.Properties["format"].(string)
	if formatDefined {
		format = fmt.Sprintf(".%s", val)
	}

	allowSchemaRelaxation, err := schemaRelaxationProperty(src.Properties)
	if err != nil {
		return err
	}

	var ingestionProps map[string]any
	if duckDBProps, ok := src.Properties["duckdb"].(map[string]any); ok {
		ingestionProps = duckDBProps
	} else {
		ingestionProps = map[string]any{}
	}
	if _, ok := ingestionProps["union_by_name"]; !ok && allowSchemaRelaxation {
		// set union_by_name to unify the schema of the files
		ingestionProps["union_by_name"] = true
	}

	allFiles := make([]string, 0)
	for iterator.HasNext() {
		files, err := iterator.NextBatch(opts.IteratorBatch)
		if err != nil {
			return err
		}
		allFiles = append(allFiles, files...)
	}

	if !formatDefined {
		format = fileutil.FullExt(allFiles[0])
	}

	st := time.Now()
	t.logger.Info("ingesting files", zap.Strings("files", allFiles), observability.ZapCtx(ctx))

	ast := opts.AST
	if ast == nil {
		query, queryDefined := src.Properties["query"].(string)
		if queryDefined && query != "" {
			ast, err = duckdbsql.Parse(query)
			if err != nil {
				return err
			}
		}
	}

	if ast != nil {
		err = ast.RewriteTableRefs(func(table *duckdbsql.TableRef) (*duckdbsql.TableRef, bool) {
			return &duckdbsql.TableRef{
				Paths:      allFiles,
				Function:   table.Function,
				Properties: table.Properties,
			}, true
		})
		if err != nil {
			return err
		}
		sql, err := ast.Format()
		if err != nil {
			return err
		}
		query := fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (%s);", dbSink.Table, sql)
		if err := t.to.Exec(ctx, &drivers.Statement{Query: query, Priority: 1}); err != nil {
			return err
		}
	} else {
		from, err := sourceReader(allFiles, format, ingestionProps)
		if err != nil {
			return err
		}

		query := fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM %s);", dbSink.Table, from)
		if err := t.to.Exec(ctx, &drivers.Statement{Query: query, Priority: 1}); err != nil {
			return err
		}
	}

	size = fileSize(allFiles)
	t.logger.Info("ingested files", zap.Strings("files", allFiles), zap.Int64("bytes_ingested", size), zap.Duration("duration", time.Since(st)), observability.ZapCtx(ctx))
	p.Observe(size, drivers.ProgressUnitByte)

	return nil
}
