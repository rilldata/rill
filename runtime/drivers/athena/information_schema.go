package athena

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

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
        WHERE T.table_schema NOT IN ('information_schema')
		%s
		ORDER BY database, database_schema, name, table_type
	`, likeClause)

	awsConfig, err := c.awsConfig(ctx, c.config.AWSRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS config: %w", err)
	}

	client := athena.NewFromConfig(awsConfig)

	queryExecutionID, err := c.executeQuery(ctx, client, q, c.config.Workgroup, c.config.OutputLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}

	result, err := client.GetQueryResults(ctx, &athena.GetQueryResultsInput{
		QueryExecutionId: aws.String(queryExecutionID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get query results: %w", err)
	}

	tables, err := scanTables(result.ResultSet.Rows)
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

	client := athena.NewFromConfig(awsConfig)

	queryExecutionID, err := c.executeQuery(ctx, client, q, c.config.Workgroup, c.config.OutputLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup table: %w", err)
	}

	result, err := client.GetQueryResults(ctx, &athena.GetQueryResultsInput{
		QueryExecutionId: aws.String(queryExecutionID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get query results: %w", err)
	}

	tables, err := scanTables(result.ResultSet.Rows)
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

func scanTables(rows []types.Row) ([]*drivers.Table, error) {
	var res []*drivers.Table

	for i, row := range rows {
		if i == 0 { // Skip header row
			continue
		}

		database := *row.Data[0].VarCharValue
		databaseSchema := *row.Data[1].VarCharValue
		isDefaultDatabaseStr := *row.Data[2].VarCharValue
		isDefaultDatabaseSchemaStr := *row.Data[3].VarCharValue
		name := *row.Data[4].VarCharValue
		tableType := *row.Data[5].VarCharValue
		columnName := *row.Data[6].VarCharValue
		columnType := *row.Data[7].VarCharValue

		isDefaultDatabase := isDefaultDatabaseStr == "true"
		isDefaultDatabaseSchema := isDefaultDatabaseSchemaStr == "true"

		// set t to res[len(res)-1] if it's the same table, else set t to a new table and append it
		var t *drivers.Table
		if len(res) > 0 {
			t = res[len(res)-1]
			if !(t.DatabaseSchema == databaseSchema && t.Name == name) {
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

var (
	errUnsupportedType = errors.New("encountered unsupported athena type")

	arrayRe = regexp.MustCompile(`^array\s*\(\s*(.+)\s*\)$`)
	mapRe   = regexp.MustCompile(`^map\s*\(\s*(.+?)\s*,\s*(.+?)\s*\)$`)
	rowRe   = regexp.MustCompile(`^row\s*\(\s*(.+)\s*\)$`)
)

func databaseTypeToPB(athenaType string) (*runtimev1.Type, error) {
	typ := strings.ToLower(strings.TrimSpace(athenaType))

	switch {
	case typ == "boolean":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_BOOL}, nil
	case typ == "tinyint":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT8}, nil
	case typ == "smallint":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT16}, nil
	case typ == "int", typ == "integer":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT32}, nil
	case typ == "bigint":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT64}, nil
	case strings.HasPrefix(typ, "decimal"), strings.HasPrefix(typ, "numeric"):
		return &runtimev1.Type{Code: runtimev1.Type_CODE_DECIMAL}, nil
	case typ == "real", typ == "float":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT32}, nil
	case typ == "double", typ == "double precision":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT64}, nil
	case strings.HasPrefix(typ, "varchar"), strings.HasPrefix(typ, "char"), typ == "string":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}, nil
	case typ == "varbinary":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_BYTES}, nil
	case typ == "json":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_JSON}, nil
	case typ == "uuid":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_UUID}, nil
	case typ == "date":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_DATE}, nil
	case typ == "time":
		return &runtimev1.Type{Code: runtimev1.Type_CODE_TIME}, nil
	case strings.HasPrefix(typ, "timestamp"):
		return &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}, nil

	case arrayRe.MatchString(typ):
		elemTypeStr := arrayRe.FindStringSubmatch(typ)[1]
		elemType, err := databaseTypeToPB(elemTypeStr)
		if err != nil {
			return nil, fmt.Errorf("array element type: %w", err)
		}
		return &runtimev1.Type{
			Code:             runtimev1.Type_CODE_ARRAY,
			ArrayElementType: elemType,
		}, nil

	case mapRe.MatchString(typ):
		matches := mapRe.FindStringSubmatch(typ)
		keyTypeStr, valTypeStr := matches[1], matches[2]
		keyType, err := databaseTypeToPB(keyTypeStr)
		if err != nil {
			return nil, fmt.Errorf("map key type: %w", err)
		}
		valType, err := databaseTypeToPB(valTypeStr)
		if err != nil {
			return nil, fmt.Errorf("map value type: %w", err)
		}
		return &runtimev1.Type{
			Code: runtimev1.Type_CODE_MAP,
			MapType: &runtimev1.MapType{
				KeyType:   keyType,
				ValueType: valType,
			},
		}, nil

	case rowRe.MatchString(typ):
		fieldsStr := rowRe.FindStringSubmatch(typ)[1]
		fields := strings.Split(fieldsStr, ",")
		var structFields []*runtimev1.StructType_Field
		for _, f := range fields {
			parts := strings.Fields(strings.TrimSpace(f))
			if len(parts) < 2 {
				return nil, fmt.Errorf("invalid struct field: %s", f)
			}
			name := parts[0]
			subType := strings.Join(parts[1:], " ")
			fieldType, err := databaseTypeToPB(subType)
			if err != nil {
				return nil, fmt.Errorf("struct field %s: %w", name, err)
			}
			structFields = append(structFields, &runtimev1.StructType_Field{
				Name: name,
				Type: fieldType,
			})
		}
		return &runtimev1.Type{
			Code: runtimev1.Type_CODE_STRUCT,
			StructType: &runtimev1.StructType{
				Fields: structFields,
			},
		}, nil

	default:
		return nil, fmt.Errorf("%w: %s", errUnsupportedType, typ)
	}
}
