package duckdb

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// OnSchemaChange controls how an incremental insert handles differences between the target and source schemas.
type OnSchemaChange string

const (
	OnSchemaChangeIgnore           OnSchemaChange = "ignore"
	OnSchemaChangeFail             OnSchemaChange = "fail"
	OnSchemaChangeAppendNewColumns OnSchemaChange = "append_new_columns"
)

func (m OnSchemaChange) Valid() bool {
	switch m {
	case OnSchemaChangeIgnore, OnSchemaChangeFail, OnSchemaChangeAppendNewColumns:
		return true
	default:
		return false
	}
}

type duckDBColumn struct {
	Name string `db:"column_name"`
	Type string `db:"data_type"`
}

// insertColumns holds the column lists to insert with, paired by position.
// The names are unquoted and must be escaped before being interpolated into SQL.
type insertColumns struct {
	target []string
	source []string
}

// hasSource reports whether name is one of the source columns being inserted.
// The comparison is case-insensitive because DuckDB identifiers are, even when quoted.
func (c *insertColumns) hasSource(name string) bool {
	for _, column := range c.source {
		if strings.EqualFold(column, name) {
			return true
		}
	}
	return false
}

// reconcileTableSchema compares the target and source schemas, applies the schema change indicated by mode,
// and returns the columns to insert with.
// Note that the ALTER TABLE statements it may issue are only rolled back on failure by rduckdb backends that mutate a copy of the table.
func reconcileTableSchema(ctx context.Context, conn *sqlx.Conn, target, source string, mode OnSchemaChange, logger *zap.Logger) (*insertColumns, error) {
	targetColumns, err := readDuckDBColumns(ctx, conn, target)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema of %q: %w", target, err)
	}
	sourceColumns, err := readDuckDBColumns(ctx, conn, source)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema of %q: %w", source, err)
	}

	targetByName := make(map[string]duckDBColumn, len(targetColumns))
	for _, column := range targetColumns {
		targetByName[strings.ToLower(column.Name)] = column
	}
	sourceByName := make(map[string]duckDBColumn, len(sourceColumns))
	for _, column := range sourceColumns {
		sourceByName[strings.ToLower(column.Name)] = column
	}

	var added, removed []duckDBColumn
	for _, sourceColumn := range sourceColumns {
		if _, ok := targetByName[strings.ToLower(sourceColumn.Name)]; !ok {
			added = append(added, sourceColumn)
		}
	}
	for _, targetColumn := range targetColumns {
		if _, ok := sourceByName[strings.ToLower(targetColumn.Name)]; !ok {
			removed = append(removed, targetColumn)
		}
	}

	switch mode {
	case OnSchemaChangeIgnore:
		if len(added) > 0 {
			logger.Warn("Discarding new columns not present in the model table",
				zap.String("table", target), zap.String("columns", formatDuckDBColumns(added)))
		}
	case OnSchemaChangeFail:
		if len(added) > 0 || len(removed) > 0 {
			return nil, fmt.Errorf(
				`the new data does not match the schema of table %q (new columns: %s; missing columns: %s). Set "on_schema_change" to allow schema changes`,
				target,
				formatDuckDBColumns(added),
				formatDuckDBColumns(removed),
			)
		}
	case OnSchemaChangeAppendNewColumns:
		for _, column := range added {
			_, err := conn.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", safeSQLName(target), safeSQLName(column.Name), column.Type))
			if err != nil {
				return nil, fmt.Errorf("failed to add column %q: %w", column.Name, err)
			}
			targetByName[strings.ToLower(column.Name)] = column
		}
		if len(added) > 0 {
			logger.Info("Added new columns to the model table",
				zap.String("table", target), zap.String("columns", formatDuckDBColumns(added)))
		}
	default:
		return nil, fmt.Errorf("invalid on_schema_change mode %q", mode)
	}

	if len(removed) > 0 && mode != OnSchemaChangeFail {
		logger.Warn("Inserting NULL for columns missing from the new data",
			zap.String("table", target), zap.String("columns", formatDuckDBColumns(removed)))
	}

	// Insert by name so that column order and any tolerated schema differences do not affect the insert.
	// Source-only columns that were not added to the target are skipped, and target-only columns are left NULL.
	cols := &insertColumns{}
	for _, sourceColumn := range sourceColumns {
		targetColumn, ok := targetByName[strings.ToLower(sourceColumn.Name)]
		if !ok {
			continue
		}
		cols.target = append(cols.target, targetColumn.Name)
		cols.source = append(cols.source, sourceColumn.Name)
	}
	if len(cols.target) == 0 {
		return nil, fmt.Errorf("the new data has no columns in common with table %q", target)
	}
	return cols, nil
}

// errTableNotFound is returned when a table does not exist on the mutation connection.
var errTableNotFound = errors.New("table not found")

// readDuckDBColumns must use the active mutation connection because rduckdb's MutateTable operates on an unpublished database copy.
// The public InformationSchema abstraction opens a separate read connection against the published table version,
// so it cannot see the regular staging table created in that copy.
func readDuckDBColumns(ctx context.Context, conn *sqlx.Conn, table string) ([]duckDBColumn, error) {
	var columns []duckDBColumn
	err := conn.SelectContext(ctx, &columns, `
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_catalog = current_database()
			AND table_schema = current_schema()
			AND lower(table_name) = lower(?)
		ORDER BY ordinal_position
	`, table)
	if err != nil {
		return nil, err
	}
	if len(columns) == 0 {
		return nil, errTableNotFound
	}
	return columns, nil
}

func formatDuckDBColumns(columns []duckDBColumn) string {
	if len(columns) == 0 {
		return "none"
	}
	formatted := make([]string, len(columns))
	for i, column := range columns {
		formatted[i] = fmt.Sprintf("%q", column.Name)
	}
	return strings.Join(formatted, ", ")
}
