package duckdb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
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

func (t *objectStoreToDuckDB) Transfer(ctx context.Context, srcProps, sinkProps map[string]any, opts *drivers.TransferOptions) error {
	sinkCfg, err := parseSinkProperties(sinkProps)
	if err != nil {
		return err
	}

	t.logger.Info("source properties", zap.Any("srcProps", srcProps))
	srcCfg, err := parseFileSourceProperties(srcProps)
	if err != nil {
		return err
	}

	iterator, err := t.from.DownloadFiles(ctx, srcProps)
	if err != nil {
		return err
	}
	defer iterator.Close()

	size, ok := iterator.Size(drivers.ProgressUnitByte)
	if ok && !sizeWithinStorageLimits(t.to, size) {
		return drivers.ErrStorageLimitExceeded
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

	a := newAppender(t.to, sinkCfg, srcCfg.AllowSchemaRelaxation, t.logger, func(files []string) (string, error) {
		from, err := sourceReader(files, format, srcCfg.DuckDB)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("SELECT * FROM %s", from), nil
	})

	for {
		files, err := iterator.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		if format == "" {
			format = fileutil.FullExt(files[0])
		}

		st := time.Now()
		t.logger.Info("ingesting files", zap.Strings("files", files), observability.ZapCtx(ctx))
		if appendToTable {
			if err := a.appendData(ctx, files); err != nil {
				return err
			}
		} else {
			from, err := sourceReader(files, format, srcCfg.DuckDB)
			if err != nil {
				return err
			}

			err = t.to.CreateTableAsSelect(ctx, sinkCfg.Table, false, fmt.Sprintf("SELECT * FROM %s", from))
			if err != nil {
				return err
			}
		}

		size := fileSize(files)
		t.logger.Info("ingested files", zap.Strings("files", files), zap.Int64("bytes_ingested", size), zap.Duration("duration", time.Since(st)), observability.ZapCtx(ctx))
		opts.Progress.Observe(size, drivers.ProgressUnitByte)
		appendToTable = true
	}
	// convert to enum
	if len(srcCfg.CastToENUM) > 0 {
		conn, _ := t.to.(*connection)
		return conn.convertToEnum(ctx, sinkCfg.Table, srcCfg.CastToENUM)
	}
	return nil
}

func (t *objectStoreToDuckDB) ingestDuckDBSQL(ctx context.Context, originalSQL string, iterator drivers.FileIterator, srcCfg *fileSourceProperties, dbSink *sinkProperties, opts *drivers.TransferOptions) error {
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
	a := newAppender(t.to, dbSink, srcCfg.AllowSchemaRelaxation, t.logger, func(files []string) (string, error) {
		return rewriteSQL(ast, files)
	})
	appendToTable := false
	for {
		files, err := iterator.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		st := time.Now()
		t.logger.Info("ingesting files", zap.Strings("files", files), observability.ZapCtx(ctx))
		if appendToTable {
			if err := a.appendData(ctx, files); err != nil {
				return err
			}
		} else {
			sql, err := rewriteSQL(ast, files)
			if err != nil {
				return err
			}

			err = t.to.CreateTableAsSelect(ctx, dbSink.Table, false, sql)
			if err != nil {
				return err
			}
		}

		size := fileSize(files)
		t.logger.Info("ingested files", zap.Strings("files", files), zap.Int64("bytes_ingested", size), zap.Duration("duration", time.Since(st)), observability.ZapCtx(ctx))
		opts.Progress.Observe(size, drivers.ProgressUnitByte)
		appendToTable = true
	}
	// convert to enum
	if len(srcCfg.CastToENUM) > 0 {
		conn, _ := t.to.(*connection)
		return conn.convertToEnum(ctx, dbSink.Table, srcCfg.CastToENUM)
	}
	return nil
}

type appender struct {
	to                    drivers.OLAPStore
	sink                  *sinkProperties
	allowSchemaRelaxation bool
	tableSchema           map[string]string
	logger                *zap.Logger
	sqlFunc               func([]string) (string, error)
}

func newAppender(to drivers.OLAPStore, sink *sinkProperties, allowSchemaRelaxation bool, logger *zap.Logger, sqlFunc func([]string) (string, error)) *appender {
	return &appender{
		to:                    to,
		sink:                  sink,
		allowSchemaRelaxation: allowSchemaRelaxation,
		logger:                logger,
		tableSchema:           nil,
		sqlFunc:               sqlFunc,
	}
}

func (a *appender) appendData(ctx context.Context, files []string) error {
	sql, err := a.sqlFunc(files)
	if err != nil {
		return err
	}

	err = a.to.InsertTableAsSelect(ctx, a.sink.Table, a.allowSchemaRelaxation, sql)
	if err == nil || !a.allowSchemaRelaxation || !containsAny(err.Error(), []string{"binder error", "conversion error"}) {
		return err
	}

	// error is of type binder error (more or less columns than current table schema)
	// or of type conversion error (datatype changed or column sequence changed)
	err = a.updateSchema(ctx, sql, files)
	if err != nil {
		return fmt.Errorf("failed to update schema %w", err)
	}
	return a.to.InsertTableAsSelect(ctx, a.sink.Table, true, sql)
}

// updateSchema updates the schema of the table in case new file adds a new column or
// updates the datatypes of an existing columns with a wider datatype.
func (a *appender) updateSchema(ctx context.Context, sql string, fileNames []string) error {
	// schema of new files
	srcSchema, err := a.scanSchemaFromQuery(ctx, fmt.Sprintf("DESCRIBE (%s);", sql))
	if err != nil {
		return err
	}

	// combined schema
	qry := fmt.Sprintf("DESCRIBE ((SELECT * FROM %s LIMIT 0) UNION ALL BY NAME (%s));", safeName(a.sink.Table), sql)
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
		if err := a.to.AddTableColumn(ctx, a.sink.Table, colName, colType); err != nil {
			return err
		}
	}

	for colName, colType := range colTypeChanged {
		a.tableSchema[colName] = colType
		if err := a.to.AlterTableColumn(ctx, a.sink.Table, colName, colType); err != nil {
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

func rewriteSQL(ast *duckdbsql.AST, allFiles []string) (string, error) {
	err := ast.RewriteTableRefs(func(table *duckdbsql.TableRef) (*duckdbsql.TableRef, bool) {
		return &duckdbsql.TableRef{
			Paths:      allFiles,
			Function:   table.Function,
			Properties: table.Properties,
			Params:     table.Params,
		}, true
	})
	if err != nil {
		return "", err
	}
	sql, err := ast.Format()
	if err != nil {
		return "", err
	}
	return sql, nil
}
