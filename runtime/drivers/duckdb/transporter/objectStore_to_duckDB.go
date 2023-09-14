package transporter

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/duckdbsql"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

const _objectStoreIteratorBatchSizeInBytes = int64(5 * datasize.GB)

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

func (t *objectStoreToDuckDB) Transfer(ctx context.Context, srcProps, sinkProps map[string]any, opts *drivers.TransferOptions) error {
	sinkCfg, err := parseSinkProperties(sinkProps)
	if err != nil {
		return err
	}

	srcCfg, err := parseFileSourceProperties(srcProps)
	if err != nil {
		return err
	}

	iterator, err := t.from.DownloadFiles(ctx, srcProps)
	if err != nil {
		return err
	}
	defer iterator.Close()

	size, _ := iterator.Size(drivers.ProgressUnitByte)
	if opts.LimitInBytes != 0 && size > opts.LimitInBytes {
		return drivers.ErrIngestionLimitExceeded
	}

	// if sql is specified use ast rewrite to fill in the downloaded files
	if srcCfg.SQL != "" {
		return t.ingestDuckDBSQL(ctx, srcCfg.SQL, iterator, srcCfg, sinkCfg, opts)
	}

	opts.Progress.Target(size, drivers.ProgressUnitByte)
	appendToTable := false
	var format string
	if srcCfg.Format != "" {
		format = fmt.Sprintf(".%s", srcCfg.Format)
	}

	if srcCfg.AllowSchemaRelaxation {
		// set union_by_name to unify the schema of the files
		srcCfg.DuckDB["union_by_name"] = true
	}

	a := newAppender(t.to, sinkCfg, srcCfg.DuckDB, srcCfg.AllowSchemaRelaxation, t.logger)

	batchSize := _objectStoreIteratorBatchSizeInBytes
	if srcCfg.BatchSizeBytes != 0 {
		batchSize = srcCfg.BatchSizeBytes
	}

	for iterator.HasNext() {
		files, err := iterator.NextBatchSize(batchSize)
		if err != nil {
			return err
		}

		if format == "" {
			format = fileutil.FullExt(files[0])
		}

		st := time.Now()
		t.logger.Info("ingesting files", zap.Strings("files", files), observability.ZapCtx(ctx))
		if appendToTable {
			if err := a.appendData(ctx, files, format); err != nil {
				return err
			}
		} else {
			from, err := sourceReader(files, format, srcCfg.DuckDB)
			if err != nil {
				return err
			}

			query := fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM %s);", safeName(sinkCfg.Table), from)
			if err := t.to.Exec(ctx, &drivers.Statement{Query: query, Priority: 1, LongRunning: true}); err != nil {
				return err
			}
		}

		size := fileSize(files)
		t.logger.Info("ingested files", zap.Strings("files", files), zap.Int64("bytes_ingested", size), zap.Duration("duration", time.Since(st)), observability.ZapCtx(ctx))
		opts.Progress.Observe(size, drivers.ProgressUnitByte)
		appendToTable = true
	}
	return nil
}

type appender struct {
	to                    drivers.OLAPStore
	sink                  *sinkProperties
	ingestionProps        map[string]any
	allowSchemaRelaxation bool
	tableSchema           map[string]string
	logger                *zap.Logger
}

func newAppender(to drivers.OLAPStore, sink *sinkProperties, ingestionProps map[string]any, allowSchemaRelaxation bool, logger *zap.Logger) *appender {
	return &appender{
		to:                    to,
		sink:                  sink,
		ingestionProps:        ingestionProps,
		allowSchemaRelaxation: allowSchemaRelaxation,
		logger:                logger,
		tableSchema:           nil,
	}
}

func (a *appender) appendData(ctx context.Context, files []string, format string) error {
	from, err := sourceReader(files, format, a.ingestionProps)
	if err != nil {
		return err
	}

	var query string
	if a.allowSchemaRelaxation {
		query = fmt.Sprintf("INSERT INTO %s BY NAME (SELECT * FROM %s);", safeName(a.sink.Table), from)
	} else {
		query = fmt.Sprintf("INSERT INTO %s (SELECT * FROM %s);", safeName(a.sink.Table), from)
	}
	a.logger.Debug("generated query", zap.String("query", query), observability.ZapCtx(ctx))
	err = a.to.Exec(ctx, &drivers.Statement{Query: query, Priority: 1, LongRunning: true})
	if err == nil || !a.allowSchemaRelaxation || !containsAny(err.Error(), []string{"binder error", "conversion error"}) {
		return err
	}

	// error is of type binder error (more or less columns than current table schema)
	// or of type conversion error (datatype changed or column sequence changed)
	err = a.updateSchema(ctx, from, files)
	if err != nil {
		return fmt.Errorf("failed to update schema %w", err)
	}

	query = fmt.Sprintf("INSERT INTO %s BY NAME (SELECT * FROM %s);", safeName(a.sink.Table), from)
	a.logger.Debug("generated query", zap.String("query", query), observability.ZapCtx(ctx))
	return a.to.Exec(ctx, &drivers.Statement{Query: query, Priority: 1, LongRunning: true})
}

// updateSchema updates the schema of the table in case new file adds a new column or
// updates the datatypes of an existing columns with a wider datatype.
func (a *appender) updateSchema(ctx context.Context, from string, fileNames []string) error {
	// schema of new files
	srcSchema, err := a.scanSchemaFromQuery(ctx, fmt.Sprintf("DESCRIBE (SELECT * FROM %s LIMIT 0);", from))
	if err != nil {
		return err
	}

	// combined schema
	qry := fmt.Sprintf("DESCRIBE ((SELECT * FROM %s limit 0) UNION ALL BY NAME (SELECT * FROM %s limit 0));", safeName(a.sink.Table), from)
	unionSchema, err := a.scanSchemaFromQuery(ctx, qry)
	if err != nil {
		return err
	}

	// current schema
	if a.tableSchema == nil {
		a.tableSchema, err = a.scanSchemaFromQuery(ctx, fmt.Sprintf("DESCRIBE %s;", safeName(a.sink.Table)))
		if err != nil {
			return err
		}
	}

	newCols := make(map[string]string)
	colTypeChanged := make(map[string]string)
	for colName, colType := range unionSchema {
		oldType, ok := a.tableSchema[colName]
		if !ok {
			newCols[colName] = colType
		} else if oldType != colType {
			colTypeChanged[colName] = colType
		}
	}

	if !a.allowSchemaRelaxation {
		if len(srcSchema) < len(unionSchema) {
			fileNames := strings.Join(names(fileNames), ",")
			columns := strings.Join(missingMapKeys(a.tableSchema, srcSchema), ",")
			return fmt.Errorf("new files %q are missing columns %q and schema relaxation not allowed", fileNames, columns)
		}

		if len(colTypeChanged) != 0 {
			fileNames := strings.Join(names(fileNames), ",")
			columns := strings.Join(keys(colTypeChanged), ",")
			return fmt.Errorf("new files %q change datatypes of some columns %q and schema relaxation not allowed", fileNames, columns)
		}
	}

	if len(newCols) != 0 && !a.allowSchemaRelaxation {
		fileNames := strings.Join(names(fileNames), ",")
		columns := strings.Join(missingMapKeys(srcSchema, a.tableSchema), ",")
		return fmt.Errorf("new files %q have new columns %q and schema relaxation not allowed", fileNames, columns)
	}

	for colName, colType := range newCols {
		a.tableSchema[colName] = colType
		qry := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", safeName(a.sink.Table), safeName(colName), colType)
		if err := a.to.Exec(ctx, &drivers.Statement{Query: qry, LongRunning: true}); err != nil {
			return err
		}
	}

	for colName, colType := range colTypeChanged {
		a.tableSchema[colName] = colType
		qry := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DATA TYPE %s", safeName(a.sink.Table), safeName(colName), colType)
		if err := a.to.Exec(ctx, &drivers.Statement{Query: qry, LongRunning: true}); err != nil {
			return err
		}
	}

	return nil
}

func (a *appender) scanSchemaFromQuery(ctx context.Context, qry string) (map[string]string, error) {
	result, err := a.to.Execute(ctx, &drivers.Statement{Query: qry, Priority: 1, LongRunning: true})
	if err != nil {
		return nil, err
	}
	defer result.Close()

	schema := make(map[string]string)
	for result.Next() {
		var s duckDBTableSchemaResult
		if err := result.StructScan(&s); err != nil {
			return nil, err
		}
		schema[s.ColumnName] = s.ColumnType
	}

	if err := result.Err(); err != nil {
		return nil, err
	}

	return schema, nil
}

func (t *objectStoreToDuckDB) ingestDuckDBSQL(ctx context.Context, originalSQL string, iterator drivers.FileIterator, srcCfg *fileSourceProperties, dbSink *sinkProperties, opts *drivers.TransferOptions) error {
	batchSize := _objectStoreIteratorBatchSizeInBytes
	if srcCfg.BatchSizeBytes != 0 {
		batchSize = srcCfg.BatchSizeBytes
	}

	iterator.KeepFilesUntilClose(true)
	allFiles := make([]string, 0)
	for iterator.HasNext() {
		files, err := iterator.NextBatchSize(batchSize)
		if err != nil {
			return err
		}
		allFiles = append(allFiles, files...)
	}

	ast, err := duckdbsql.Parse(originalSQL)
	if err != nil {
		return err
	}

	// Validate the sql is supported for sources
	// TODO: find a good common place for this validation and avoid code duplication here and in sources packages as well
	refs := ast.GetTableRefs()
	if len(refs) != 1 {
		return errors.New("sql source should have exactly one table reference")
	}
	ref := refs[0]

	if len(ref.Paths) == 0 {
		return errors.New("only read_* functions with a single path is supported")
	}
	if len(ref.Paths) > 1 {
		return errors.New("invalid source, only a single path for source is supported")
	}

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

	st := time.Now()
	query := fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (%s);", dbSink.Table, sql)
	err = t.to.Exec(ctx, &drivers.Statement{Query: query, Priority: 1, LongRunning: true})
	if err != nil {
		return err
	}

	size := fileSize(allFiles)
	t.logger.Info("ingested files", zap.Strings("files", allFiles), zap.Int64("bytes_ingested", size), zap.Duration("duration", time.Since(st)), observability.ZapCtx(ctx))
	opts.Progress.Observe(size, drivers.ProgressUnitByte)
	return nil
}
