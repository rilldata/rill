package transporter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
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
	appendToTable := false
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

	a := newAppender(t.to, dbSink, ingestionProps, allowSchemaRelaxation, t.logger)

	for iterator.HasNext() {
		files, err := iterator.NextBatch(opts.IteratorBatch)
		if err != nil {
			return err
		}

		if !formatDefined {
			format = fileutil.FullExt(files[0])
			formatDefined = true
		}

		st := time.Now()
		t.logger.Info("ingesting files", zap.Strings("files", files), observability.ZapCtx(ctx))
		if appendToTable {
			if err := a.appendData(ctx, files, format); err != nil {
				return err
			}
		} else {
			from, err := sourceReader(files, format, ingestionProps)
			if err != nil {
				return err
			}

			query := fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM %s);", dbSink.Table, from)
			if err := t.to.Exec(ctx, &drivers.Statement{Query: query, Priority: 1}); err != nil {
				return err
			}
		}

		size := fileSize(files)
		t.logger.Info("ingested files", zap.Strings("files", files), zap.Int64("bytes_ingested", size), zap.Duration("duration", time.Since(st)), observability.ZapCtx(ctx))
		p.Observe(size, drivers.ProgressUnitByte)
		appendToTable = true
	}
	return nil
}

type appender struct {
	to                    drivers.OLAPStore
	sink                  *drivers.DatabaseSink
	ingestionProps        map[string]any
	allowSchemaRelaxation bool
	tableSchema           map[string]string
	logger                *zap.Logger
}

func newAppender(to drivers.OLAPStore, sink *drivers.DatabaseSink, ingestionProps map[string]any,
	allowSchemaRelaxation bool, logger *zap.Logger,
) *appender {
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
		query = fmt.Sprintf("INSERT INTO %q BY NAME (SELECT * FROM %s);", a.sink.Table, from)
	} else {
		query = fmt.Sprintf("INSERT INTO %q (SELECT * FROM %s);", a.sink.Table, from)
	}
	a.logger.Debug("generated query", zap.String("query", query), observability.ZapCtx(ctx))
	err = a.to.Exec(ctx, &drivers.Statement{Query: query, Priority: 1})
	if err == nil || !a.allowSchemaRelaxation || !containsAny(err.Error(), []string{"binder error", "conversion error"}) {
		return err
	}

	// error is of type binder error (more or less columns than current table schema)
	// or of type conversion error (datatype changed or column sequence changed)
	err = a.updateSchema(ctx, from, files)
	if err != nil {
		return fmt.Errorf("failed to update schema %w", err)
	}

	query = fmt.Sprintf("INSERT INTO %q BY NAME (SELECT * FROM %s);", a.sink.Table, from)
	a.logger.Debug("generated query", zap.String("query", query), observability.ZapCtx(ctx))
	return a.to.Exec(ctx, &drivers.Statement{Query: query, Priority: 1})
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
	qry := fmt.Sprintf("DESCRIBE ((SELECT * FROM %s limit 0) UNION ALL BY NAME (SELECT * FROM %s limit 0));", a.sink.Table, from)
	unionSchema, err := a.scanSchemaFromQuery(ctx, qry)
	if err != nil {
		return err
	}

	// current schema
	if a.tableSchema == nil {
		a.tableSchema, err = a.scanSchemaFromQuery(ctx, fmt.Sprintf("DESCRIBE %q;", a.sink.Table))
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
		qry := fmt.Sprintf("ALTER TABLE %q ADD COLUMN %q %s", a.sink.Table, colName, colType)
		if err := a.to.Exec(ctx, &drivers.Statement{Query: qry}); err != nil {
			return err
		}
	}

	for colName, colType := range colTypeChanged {
		a.tableSchema[colName] = colType
		qry := fmt.Sprintf("ALTER TABLE %q ALTER COLUMN %q SET DATA TYPE %s", a.sink.Table, colName, colType)
		if err := a.to.Exec(ctx, &drivers.Statement{Query: qry}); err != nil {
			return err
		}
	}

	return nil
}

func (a *appender) scanSchemaFromQuery(ctx context.Context, qry string) (map[string]string, error) {
	result, err := a.to.Execute(ctx, &drivers.Statement{Query: qry, Priority: 1})
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
