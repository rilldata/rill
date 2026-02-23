package oracle

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func TestResolveDSN_WithOnlyDSN(t *testing.T) {
	c := &ConfigProperties{
		DSN: "oracle://user:pass@localhost:1521/ORCLPDB1",
	}

	dsn, err := c.ResolveDSN()
	require.NoError(t, err)
	require.Equal(t, c.DSN, dsn)
}

func TestResolveDSN_WithIndividualFields(t *testing.T) {
	c := &ConfigProperties{
		Host:        "db.example.com",
		Port:        1521,
		User:        "admin",
		Password:    "secret",
		ServiceName: "ORCLPDB1",
	}

	dsn, err := c.ResolveDSN()
	require.NoError(t, err)
	require.Equal(t, "oracle://admin:secret@db.example.com:1521/ORCLPDB1", dsn)
}

func TestResolveDSN_WithDefaults(t *testing.T) {
	c := &ConfigProperties{
		User:        "admin",
		ServiceName: "ORCL",
	}

	dsn, err := c.ResolveDSN()
	require.NoError(t, err)
	require.Equal(t, "oracle://admin@localhost:1521/ORCL", dsn)
}

func TestResolveDSN_WithPasswordSpecialChars(t *testing.T) {
	c := &ConfigProperties{
		Host:        "localhost",
		Port:        1521,
		User:        "admin",
		Password:    "p@ss:w0rd/test",
		ServiceName: "ORCL",
	}

	dsn, err := c.ResolveDSN()
	require.NoError(t, err)
	require.Contains(t, dsn, "oracle://admin:")
	require.Contains(t, dsn, "@localhost:1521/ORCL")
}

func TestResolveDSN_WithDSNAndIndividualFields_ShouldError(t *testing.T) {
	c := &ConfigProperties{
		DSN:  "oracle://some:dsn@localhost/db",
		Host: "localhost",
	}

	_, err := c.ResolveDSN()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid config")
}

func TestResolveDSN_DSNConflictWithUser(t *testing.T) {
	c := &ConfigProperties{
		DSN:  "oracle://user:pass@localhost:1521/ORCL",
		User: "admin",
	}

	_, err := c.ResolveDSN()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid config")
}

func TestDatabaseTypeToPB(t *testing.T) {
	tests := []struct {
		dbType   string
		expected runtimev1.Type_Code
	}{
		{"NUMBER", runtimev1.Type_CODE_FLOAT64},
		{"FLOAT", runtimev1.Type_CODE_FLOAT32},
		{"BINARY_FLOAT", runtimev1.Type_CODE_FLOAT32},
		{"BINARY_DOUBLE", runtimev1.Type_CODE_FLOAT64},
		{"INTEGER", runtimev1.Type_CODE_INT64},
		{"INT", runtimev1.Type_CODE_INT64},
		{"SMALLINT", runtimev1.Type_CODE_INT64},
		{"VARCHAR2", runtimev1.Type_CODE_STRING},
		{"NVARCHAR2", runtimev1.Type_CODE_STRING},
		{"CHAR", runtimev1.Type_CODE_STRING},
		{"NCHAR", runtimev1.Type_CODE_STRING},
		{"CLOB", runtimev1.Type_CODE_STRING},
		{"NCLOB", runtimev1.Type_CODE_STRING},
		{"LONG", runtimev1.Type_CODE_STRING},
		{"ROWID", runtimev1.Type_CODE_STRING},
		{"BLOB", runtimev1.Type_CODE_BYTES},
		{"RAW", runtimev1.Type_CODE_BYTES},
		{"LONG RAW", runtimev1.Type_CODE_BYTES},
		{"DATE", runtimev1.Type_CODE_TIMESTAMP},
		{"TIMESTAMP", runtimev1.Type_CODE_TIMESTAMP},
		{"TIMESTAMP WITH TIME ZONE", runtimev1.Type_CODE_TIMESTAMP},
		{"TIMESTAMP WITH LOCAL TIME ZONE", runtimev1.Type_CODE_TIMESTAMP},
		{"BOOLEAN", runtimev1.Type_CODE_BOOL},
		{"JSON", runtimev1.Type_CODE_JSON},
		{"XMLTYPE", runtimev1.Type_CODE_STRING},
		{"INTERVAL YEAR TO MONTH", runtimev1.Type_CODE_STRING},
		{"INTERVAL DAY TO SECOND", runtimev1.Type_CODE_STRING},
		{"UNKNOWN_TYPE", runtimev1.Type_CODE_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.dbType, func(t *testing.T) {
			result := databaseTypeToPB(tt.dbType, true)
			require.Equal(t, tt.expected, result.Code)
			require.True(t, result.Nullable)
		})
	}
}

func TestDatabaseTypeToPB_NotNullable(t *testing.T) {
	result := databaseTypeToPB("VARCHAR2", false)
	require.Equal(t, runtimev1.Type_CODE_STRING, result.Code)
	require.False(t, result.Nullable)
}
