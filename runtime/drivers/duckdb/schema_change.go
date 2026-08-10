package duckdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
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

type schemaChangePlan struct {
	add              []duckDBColumn
	targetInsertCols []string
	sourceSelectCols []string
}

func reconcileTableSchema(ctx context.Context, conn *sqlx.Conn, target, source string, mode OnSchemaChange) (*schemaChangePlan, error) {
	targetColumns, err := readDuckDBColumns(ctx, conn, target)
	if err != nil {
		return nil, fmt.Errorf("failed to read target schema: %w", err)
	}
	sourceColumns, err := readDuckDBColumns(ctx, conn, source)
	if err != nil {
		return nil, fmt.Errorf("failed to read source schema: %w", err)
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

	plan := &schemaChangePlan{}
	switch mode {
	case OnSchemaChangeIgnore:
		if len(removed) > 0 {
			return nil, fmt.Errorf("schema change mode %q does not allow target columns to be absent from the source: %s", mode, formatDuckDBColumns(removed))
		}
	case OnSchemaChangeFail:
		if len(added) > 0 || len(removed) > 0 {
			return nil, fmt.Errorf(
				"schema change mode %q detected schema differences (added: %s; removed: %s)",
				mode,
				formatDuckDBColumns(added),
				formatDuckDBColumns(removed),
			)
		}
	case OnSchemaChangeAppendNewColumns:
		plan.add = added
	default:
		return nil, fmt.Errorf("invalid on_schema_change mode %q", mode)
	}

	for _, sourceColumn := range sourceColumns {
		targetColumn, ok := targetByName[strings.ToLower(sourceColumn.Name)]
		if ok {
			plan.targetInsertCols = append(plan.targetInsertCols, targetColumn.Name)
			plan.sourceSelectCols = append(plan.sourceSelectCols, sourceColumn.Name)
			continue
		}
		if mode != OnSchemaChangeIgnore {
			plan.targetInsertCols = append(plan.targetInsertCols, sourceColumn.Name)
			plan.sourceSelectCols = append(plan.sourceSelectCols, sourceColumn.Name)
		}
	}

	for _, column := range plan.add {
		_, err := conn.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", safeSQLName(target), safeSQLName(column.Name), column.Type))
		if err != nil {
			return nil, fmt.Errorf("failed to add column %q: %w", column.Name, err)
		}
	}
	return plan, nil
}

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
		return nil, fmt.Errorf("table %q has no columns", table)
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
