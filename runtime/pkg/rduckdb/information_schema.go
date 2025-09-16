package rduckdb

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pagination"
)

type Table struct {
	Database       string `db:"database"`
	Schema         string `db:"schema"`
	Name           string `db:"name"`
	View           bool   `db:"view"`
	ColumnNames    []any  `db:"column_names"`
	ColumnTypes    []any  `db:"column_types"`
	ColumnNullable []any  `db:"column_nullable"`
	SizeBytes      int64  `db:"-"`
}

func (d *db) Schema(ctx context.Context, ilike, name string, pageSize uint32, pageToken string) ([]*Table, string, error) {
	if ilike != "" && name != "" {
		return nil, "", fmt.Errorf("cannot specify both `ilike` and `name`")
	}
	connx, release, err := d.AcquireReadConnection(ctx)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		_ = release()
	}()

	var whereClause string
	var args []any
	if ilike != "" {
		whereClause = " AND t.table_name ilike ?"
		args = []any{ilike}
	} else if name != "" {
		whereClause = " AND t.table_name = ?"
		args = []any{name}
	}

	// Add pagination filter
	if pageToken != "" {
		var startAfterName string
		if err := pagination.UnmarshalPageToken(pageToken, &startAfterName); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		whereClause += " AND t.table_name > ?"
		args = append(args, startAfterName)
	}

	// the database schema changes with every ingestion
	// we pin the read connection to the latest schema and set schema as `main` to give impression that everything is in the same schema
	q := fmt.Sprintf(`
		SELECT
			coalesce(t.table_catalog, current_database()) AS "database",
			'main' AS "schema",
			t.table_name AS "name",
			t.table_type = 'VIEW' AS "view", 
			array_agg(c.column_name ORDER BY c.ordinal_position) AS "column_names",
			array_agg(c.data_type ORDER BY c.ordinal_position) AS "column_types",
			array_agg(c.is_nullable = 'YES' ORDER BY c.ordinal_position) AS "column_nullable"
		FROM information_schema.tables t
		JOIN information_schema.columns c 
			ON t.table_schema = c.table_schema 
			AND t.table_name = c.table_name
		WHERE database = current_database() 
			AND t.table_schema = current_schema()
			%s
		GROUP BY ALL
		ORDER BY t.table_name
		LIMIT ?
	`, whereClause)

	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)
	args = append(args, limit+1)

	var res []*Table
	err = connx.SelectContext(ctx, &res, q, args...)
	if err != nil {
		return nil, "", err
	}

	// due to external table storage the information_schema always returns table type as view
	// so we look at catalog to determine if it is a view or table
	// NOTE : there is a chance of inconsistency since tables can get updated between these two calls
	tables := d.catalog.listTables()
	catalog := make(map[string]*tableMeta)
	for _, table := range tables {
		catalog[table.Name] = table
	}
	for _, t := range res {
		table, ok := catalog[t.Name]
		if ok {
			t.View = table.Type == "VIEW"
			if !t.View {
				t.SizeBytes = fileSize([]string{d.localDBPath(t.Name, table.Version)})
			}
		}
	}

	next := ""
	if len(res) > limit {
		res = res[:limit]
		next = pagination.MarshalPageToken(res[len(res)-1].Name)
	}

	return res, next, nil
}
