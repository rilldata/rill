package druid

import (
	"context"
	"fmt"
	"strings"

	_ "github.com/apache/calcite-avatica-go/v5"
	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/infra"
)

func init() {
	infra.Register("druid", driver{})
}

type driver struct{}

// Open connects to a Druid cluster using Avatica. Note that the Druid connection string must have
// the form "http://host/druid/v2/sql/avatica-protobuf/".
func (d driver) Open(dsn string) (infra.Connection, error) {
	db, err := sqlx.Open("avatica", dsn)
	if err != nil {
		return nil, err
	}

	conn := &connection{db: db}
	return conn, nil
}

type connection struct {
	db *sqlx.DB
}

func (c *connection) Close() error {
	return c.db.Close()
}

type informationSchema struct {
	conn *connection
}

func (c *connection) InformationSchema() infra.InformationSchema {
	return &informationSchema{conn: c}
}

func (c *connection) Execute(ctx context.Context, stmt *infra.Statement) (*sqlx.Rows, error) {
	if stmt.DryRun {
		// TODO: Find way to validate with args
		prepared, err := c.db.PrepareContext(ctx, stmt.Query)
		if err != nil {
			return nil, err
		}
		prepared.Close()
		return nil, nil
	}

	rows, err := c.db.QueryxContext(ctx, stmt.Query, stmt.Args...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}


func (is informationSchema) All() ([]*infra.Table, error) {
	qry := fmt.Sprintf(`SELECT t.TABLE_CATALOG  as "Database", t.TABLE_SCHEMA as "Schema", t.TABLE_NAME as "Name", t.TABLE_TYPE as "Type", 
	c.COLUMN_NAME as "Columns", c.DATA_TYPE as "ColumnType" FROM INFORMATION_SCHEMA.TABLES t 
	join INFORMATION_SCHEMA.COLUMNS c on t.TABLE_SCHEMA = c.TABLE_SCHEMA AND t.TABLE_NAME = c.TABLE_NAME`)
	table, err := getAggregatedSchema(is, qry)
	if err != nil {
		return nil, err
	}
	return table, nil
}

func (is informationSchema) Lookup(name string) (*infra.Table, error) {
	qry := fmt.Sprintf(`SELECT t.TABLE_CATALOG  as "Database", t.TABLE_SCHEMA as "Schema", t.TABLE_NAME as "Name", t.TABLE_TYPE as "Type", 
	c.COLUMN_NAME as "Columns", c.DATA_TYPE as "ColumnType" FROM INFORMATION_SCHEMA.TABLES t 
	join INFORMATION_SCHEMA.COLUMNS c on t.TABLE_SCHEMA = c.TABLE_SCHEMA AND t.TABLE_NAME = c.TABLE_NAME 
	WHERE t.TABLE_NAME = '%s' `, name)
	table, err := getAggregatedSchema(is,qry)
	if err != nil {
		return nil, err
	}

	if len(table) == 0 {
		return nil, fmt.Errorf("Table not Found")
	}
	return table[0], nil
}

func getAggregatedSchema(is informationSchema, qry string) ([]*infra.Table, error) {
	rows, err := is.conn.Execute(context.Background(), &infra.Statement{Query: qry})
	if err != nil {
		return nil, err
	}

	res := map[schemaKey][]string{}
	var table []*infra.Table
	result := schemaResults{}

	for rows.Next() {
		err := rows.Scan(&result.Database, &result.Schema, &result.Name, &result.Type, &result.ColumnName, &result.ColumnType)
		if err != nil {
			return nil, err
		}
		key := schemaKey{result.Database, result.Schema, result.Name, result.Type}
		res[key] = append(res[key], result.ColumnName+"$"+result.ColumnType)
	}
	
	for key, elements := range res {
		var info infra.Table
		info.Database = key.Database
		info.Name = key.Name
		info.Schema = key.Schema
		info.Type = key.Type

		var columns []infra.Column
		for _, element := range elements {

			var column infra.Column
			cols := strings.Split(element, "$")
			column.Name = cols[0]
			column.Type = cols[1]
			columns = append(columns, column)
		}
		info.Columns = columns
		table = append(table, &info)

	}

	return table, nil
}

type schemaKey struct {
	Database string
	Schema   string
	Name     string
	Type     string
}

type schemaResults struct {
	Database   string
	Schema     string
	Name       string
	Type       string
	ColumnName string
	ColumnType string
}

