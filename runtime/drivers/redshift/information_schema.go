package redshift

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata"
	"github.com/aws/aws-sdk-go-v2/service/redshiftdata/types"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

var errUnsupportedType = errors.New("encountered unsupported redshift type")

func (c *Connection) ListDatabases(ctx context.Context) ([]string, error) {
	q := `SHOW DATABASES`

	awsConfig, err := c.awsConfig(ctx, c.config.AWSRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS config: %w", err)
	}

	client := redshiftdata.NewFromConfig(awsConfig)

	queryExecutionID, err := c.executeQuery(ctx, client, q, c.config.Database, c.config.Workgroup, c.config.ClusterIdentifier)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}

	result, err := client.GetStatementResult(ctx, &redshiftdata.GetStatementResultInput{
		Id: aws.String(queryExecutionID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get query results: %w", err)
	}

	var names []string
	for i, record := range result.Records {
		if i == 0 { // Skip header row
			continue
		}
		name := record[0].(*types.FieldMemberStringValue).Value
		names = append(names, name)
	}

	return names, nil
}

func (c *Connection) ListSchemas(ctx context.Context, database string) ([]string, error) {
	q := fmt.Sprintf(`SHOW SCHEMAS FROM DATABASE %s`, database)

	awsConfig, err := c.awsConfig(ctx, c.config.AWSRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS config: %w", err)
	}

	client := redshiftdata.NewFromConfig(awsConfig)

	queryExecutionID, err := c.executeQuery(ctx, client, q, c.config.Database, c.config.Workgroup, c.config.ClusterIdentifier)
	if err != nil {
		return nil, fmt.Errorf("failed to list schemas: %w", err)
	}

	result, err := client.GetStatementResult(ctx, &redshiftdata.GetStatementResultInput{
		Id: aws.String(queryExecutionID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get query results: %w", err)
	}

	var names []string
	exclude := map[string]bool{
		"information_schema": true,
		"pg_catalog":         true,
		"pg_internal":        true,
	}
	for i, record := range result.Records {
		if i == 0 { // Skip header row
			continue
		}
		name := record[1].(*types.FieldMemberStringValue).Value
		if !exclude[name] {
			names = append(names, name)
		}
	}

	return names, nil
}

func (c *Connection) All(ctx context.Context, like string) ([]*drivers.Table, error) {
	var likeClause string
	if like != "" {
		likeClause = "AND (LOWER(T.table_name) LIKE LOWER('" + like + "') OR CONCAT(T.table_schema, '.', T.table_name) LIKE LOWER('" + like + "'))"
	}

	q := fmt.Sprintf(`
		SELECT 
		    T.table_catalog AS database,
			T.table_schema AS database_schema,
			true AS is_default_database,
			true AS is_default_database_schema,
			T.table_name AS name,
			T.table_type,
			C.column_name AS columns,
			C.data_type AS column_type
		FROM information_schema.tables T
		JOIN information_schema.columns C ON T.table_catalog = C.table_catalog AND T.table_schema = C.table_schema AND T.table_name = C.table_name
		WHERE T.table_schema NOT IN ('information_schema', 'pg_catalog', 'pg_auto_copy')
		%s
		ORDER BY database, database_schema, name, table_type
	`, likeClause)

	awsConfig, err := c.awsConfig(ctx, c.config.AWSRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS config: %w", err)
	}

	client := redshiftdata.NewFromConfig(awsConfig)

	queryExecutionID, err := c.executeQuery(ctx, client, q, c.config.Database, c.config.Workgroup, c.config.ClusterIdentifier)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}

	result, err := client.GetStatementResult(ctx, &redshiftdata.GetStatementResultInput{
		Id: aws.String(queryExecutionID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get query results: %w", err)
	}

	tables, err := c.scanTables(result.Records)
	if err != nil {
		return nil, err
	}
	return tables, nil
}

func (c *Connection) Lookup(ctx context.Context, db, schema, name string) (*drivers.Table, error) {
	q := fmt.Sprintf(`
		SELECT 
			T.table_catalog AS database,
			T.table_schema AS database_schema,
			true AS is_default_database,
			true AS is_default_database_schema,
			T.table_name AS name,
			T.table_type,
			C.column_name AS columns,
			C.data_type AS column_type
		FROM information_schema.tables T
		JOIN information_schema.columns C ON T.table_catalog = C.table_catalog AND T.table_schema = C.table_schema AND T.table_name = C.table_name
		WHERE T.table_catalog = '%s' AND T.table_schema = '%s' AND T.table_name = '%s'
		ORDER BY database, database_schema, name, table_type, ordinal_position
	`, db, schema, name)

	awsConfig, err := c.awsConfig(ctx, c.config.AWSRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS config: %w", err)
	}

	client := redshiftdata.NewFromConfig(awsConfig)

	queryExecutionID, err := c.executeQuery(ctx, client, q, c.config.Database, c.config.Workgroup, c.config.ClusterIdentifier)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup table: %w", err)
	}

	result, err := client.GetStatementResult(ctx, &redshiftdata.GetStatementResultInput{
		Id: aws.String(queryExecutionID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get query results: %w", err)
	}

	tables, err := c.scanTables(result.Records)
	if err != nil {
		return nil, err
	}

	if len(tables) == 0 {
		return nil, drivers.ErrNotFound
	}

	return tables[0], nil
}

func (c *Connection) LoadPhysicalSize(ctx context.Context, tables []*drivers.Table) error {
	if len(tables) == 0 {
		return nil
	}
	for _, t := range tables {
		t.PhysicalSizeBytes = -1
	}
	return nil
}

func (c *Connection) scanTables(records [][]types.Field) ([]*drivers.Table, error) {
	var res []*drivers.Table

	for i, record := range records {
		if i == 0 { // Skip header row
			continue
		}

		database := record[0].(*types.FieldMemberStringValue).Value
		databaseSchema := record[1].(*types.FieldMemberStringValue).Value
		isDefaultDatabase := record[2].(*types.FieldMemberBooleanValue).Value
		isDefaultDatabaseSchema := record[3].(*types.FieldMemberBooleanValue).Value
		name := record[4].(*types.FieldMemberStringValue).Value
		tableType := record[5].(*types.FieldMemberStringValue).Value
		columnName := record[6].(*types.FieldMemberStringValue).Value
		columnType := record[7].(*types.FieldMemberStringValue).Value

		// set t to res[len(res)-1] if it's the same table, else set t to a new table and append it
		var t *drivers.Table
		if len(res) > 0 {
			t = res[len(res)-1]
			if !(t.Database == database && t.DatabaseSchema == databaseSchema && t.Name == name) {
				t = nil
			}
		}
		if t == nil {
			t = &drivers.Table{
				Database:                database,
				DatabaseSchema:          databaseSchema,
				IsDefaultDatabase:       isDefaultDatabase,
				IsDefaultDatabaseSchema: isDefaultDatabaseSchema,
				Name:                    name,
				View:                    tableType == "VIEW",
				Schema:                  &runtimev1.StructType{},
				PhysicalSizeBytes:       -1,
			}
			res = append(res, t)
		}

		// parse column type
		colType, err := databaseTypeToPB(columnType)
		if err != nil {
			if !errors.Is(err, errUnsupportedType) {
				return nil, err
			}
			if t.UnsupportedCols == nil {
				t.UnsupportedCols = make(map[string]string)
			}
			t.UnsupportedCols[columnName] = columnType
			continue
		}

		// append column
		t.Schema.Fields = append(t.Schema.Fields, &runtimev1.StructType_Field{
			Name: columnName,
			Type: colType,
		})
	}

	return res, nil
}

func databaseTypeToPB(redshiftType string) (*runtimev1.Type, error) {
	typ := strings.ToLower(strings.TrimSpace(redshiftType))

	switch {
	case typ == "boolean":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_BOOL}, nil
	case typ == "smallint":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT16}, nil
	case typ == "int", typ == "integer":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT32}, nil
	case typ == "bigint":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT64}, nil
	case strings.HasPrefix(typ, "decimal"), strings.HasPrefix(typ, "numeric"):
		return &runtimev1.Type{Code: runtimev1.Type_CODE_DECIMAL}, nil
	case typ == "real", typ == "float4":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT32}, nil
	case typ == "double precision", typ == "float8":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT64}, nil
	case strings.HasPrefix(typ, "varchar"), strings.HasPrefix(typ, "char"), typ == "text":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}, nil
	case typ == "bytea":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_BYTES}, nil
	case typ == "json", typ == "jsonb":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_JSON}, nil
	case typ == "uuid":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_UUID}, nil
	case typ == "date":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_DATE}, nil
	case strings.HasPrefix(typ, "timestamp"):
		return &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}, nil
	case typ == "time", strings.HasPrefix(typ, "time"):
		return &runtimev1.Type{Code: runtimev1.Type_CODE_TIME}, nil
	case typ == "intervald2s", typ == "intervaly2m":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INTERVAL}, nil
	case typ == "super":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_JSON}, nil
	default:
		return nil, fmt.Errorf("%w: %s", errUnsupportedType, typ)
	}
}
