package duckdb

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/marcboeker/go-duckdb"
	"github.com/rilldata/rill/runtime/infra"
	"github.com/rilldata/rill/runtime/pkg/priorityworker"
)

func init() {
	infra.Register("duckdb", driver{})
}

type driver struct{}

func (d driver) Open(dsn string) (infra.Connection, error) {
	db, err := sqlx.Open("duckdb", dsn)
	if err != nil {
		return nil, err
	}

	conn := &connection{db: db}
	conn.worker = priorityworker.New(conn.executeQuery)

	return conn, nil
}

type connection struct {
	db     *sqlx.DB
	worker *priorityworker.PriorityWorker[*job]
}

type job struct {
	stmt   *infra.Statement
	result *sqlx.Rows
}

type informationSchema struct {
	conn *connection
}

func (c *connection) InformationSchema() infra.InformationSchema {
	return &informationSchema{conn: c}
}

type Map map[any]any

func (is informationSchema) All() ([]*infra.Table, error) {
	qry := `SELECT CASE WHEN t.TABLE_CATALOG IS NULL THEN '' ELSE t.TABLE_CATALOG END  as "Database", t.TABLE_SCHEMA as "Schema", t.TABLE_NAME as "Name", t.TABLE_TYPE as "Type", 
	map(ARRAY_AGG(c.COLUMN_NAME), ARRAY_AGG(c.DATA_TYPE)) as "Columns" FROM INFORMATION_SCHEMA.TABLES t join INFORMATION_SCHEMA.COLUMNS c 
	ON t.TABLE_SCHEMA = c.TABLE_SCHEMA AND t.TABLE_NAME = c.TABLE_NAME GROUP BY 1,2,3,4`
	table, err := getInformationSchema(is, qry)
	if err != nil {
		return nil, err
	}

	return table, nil
}

func (is informationSchema) Lookup(tableName string) (*infra.Table, error) {
	qry := fmt.Sprintf(`SELECT CASE WHEN t.TABLE_CATALOG IS NULL THEN '' ELSE t.TABLE_CATALOG END  as "Database", t.TABLE_SCHEMA as "Schema", t.TABLE_NAME as "Name", t.TABLE_TYPE as "Type", 
	map(ARRAY_AGG(c.COLUMN_NAME), ARRAY_AGG(c.DATA_TYPE)) as "Columns" FROM INFORMATION_SCHEMA.TABLES t join INFORMATION_SCHEMA.COLUMNS c 
	ON t.TABLE_SCHEMA = c.TABLE_SCHEMA AND t.TABLE_NAME = c.TABLE_NAME WHERE t.TABLE_NAME = '%s' GROUP BY 1,2,3,4 `, tableName)

	table, err := getInformationSchema(is, qry)
	if err != nil {
		return nil, err
	}

	if len(table) == 0 {
		return nil, fmt.Errorf("Table not Found")
	}
	
	return table[0], nil
}

func getInformationSchema(is informationSchema, qry string) ([]*infra.Table, error) {
	rows, err := is.conn.Execute(context.Background(), &infra.Statement{Query: qry})
	if err != nil {
		return nil, err
	}

	var tables []*infra.Table
	for rows.Next() {
		var table infra.Table
		var colsMap Map
		err = rows.Scan(&table.Database, &table.Schema, &table.Name, &table.Type, &colsMap)
		if err != nil {
			return nil, err
		}

		var columns []infra.Column
		for key, value := range colsMap {
			var column infra.Column

			column.Name = key.(string)
			column.Type = value.(string)

			columns = append(columns, column)
		}

		table.Columns = columns
		tables = append(tables, &table)
	}

	return tables, nil

}

func (c *connection) Execute(ctx context.Context, stmt *infra.Statement) (*sqlx.Rows, error) {
	j := &job{
		stmt: stmt,
	}

	err := c.worker.Process(ctx, stmt.Priority, j)
	if err != nil {
		if err == priorityworker.ErrStopped {
			return nil, infra.ErrClosed
		}
		return nil, err
	}

	return j.result, nil
}

func (c *connection) executeQuery(ctx context.Context, j *job) error {
	if j.stmt.DryRun {
		// TODO: Find way to validate with args
		prepared, err := c.db.PrepareContext(ctx, j.stmt.Query)
		if err != nil {
			return err
		}
		prepared.Close()
		return nil
	}

	rows, err := c.db.QueryxContext(ctx, j.stmt.Query, j.stmt.Args...)
	j.result = rows
	return err
}

func (c *connection) Close() error {
	c.worker.Stop()
	return c.db.Close()
}
