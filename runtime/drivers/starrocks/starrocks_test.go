package starrocks

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigPropertiesValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *ConfigProperties
		wantErr bool
	}{
		{
			name:    "empty config",
			cfg:     &ConfigProperties{},
			wantErr: true,
		},
		{
			name: "dsn only",
			cfg: &ConfigProperties{
				DSN: "user:pass@tcp(host:9030)/db",
			},
			wantErr: false,
		},
		{
			name: "host only",
			cfg: &ConfigProperties{
				Host: "localhost",
			},
			wantErr: false,
		},
		{
			name: "both dsn and host",
			cfg: &ConfigProperties{
				DSN:  "user:pass@tcp(host:9030)/db",
				Host: "localhost",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *ConfigProperties
		contains string // substring that should be in the result
	}{
		{
			name: "dsn passthrough",
			cfg: &ConfigProperties{
				DSN: "user:pass@tcp(host:9030)/db?parseTime=true",
			},
			contains: "user:pass@tcp(host:9030)/db",
		},
		{
			name: "build from fields",
			cfg: &ConfigProperties{
				Host:     "localhost",
				Port:     9030,
				Username: "root",
				Password: "secret",
			},
			contains: "root:secret@tcp(localhost:9030)",
		},
		{
			name: "build from fields with ssl",
			cfg: &ConfigProperties{
				Host:     "localhost",
				Port:     9030,
				Username: "root",
				SSL:      true,
			},
			contains: "tls=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &connection{configProp: tt.cfg}
			result := c.buildDSN()
			require.Contains(t, result, tt.contains)
		})
	}
}

func TestSafeSQLName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "table_name",
			expected: "`table_name`",
		},
		{
			name:     "name with backtick",
			input:    "table`name",
			expected: "`table``name`",
		},
		{
			name:     "empty name",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := safeSQLName(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestDatabaseTypeToRuntimeType(t *testing.T) {
	c := &connection{}

	tests := []struct {
		dbType    string
		expected  string
		expectErr bool
	}{
		// Boolean types
		{"BOOLEAN", "CODE_BOOL", false},
		{"BOOL", "CODE_BOOL", false},

		// Integer types
		{"TINYINT", "CODE_INT8", false},
		{"SMALLINT", "CODE_INT16", false},
		{"INT", "CODE_INT32", false},
		{"INTEGER", "CODE_INT32", false},
		{"BIGINT", "CODE_INT64", false},
		{"LARGEINT", "CODE_INT128", false},

		// Floating point types
		{"FLOAT", "CODE_FLOAT32", false},
		{"DOUBLE", "CODE_FLOAT64", false},
		{"DECIMAL(10,2)", "CODE_STRING", false}, // DECIMAL returns string for precision

		// String types
		{"CHAR(10)", "CODE_STRING", false},
		{"VARCHAR(255)", "CODE_STRING", false},
		{"STRING", "CODE_STRING", false},
		{"TEXT", "CODE_STRING", false},

		// Binary types (same as MySQL - returns CODE_STRING)
		{"BINARY", "CODE_STRING", false},
		{"VARBINARY", "CODE_STRING", false},
		{"BLOB", "CODE_STRING", false},

		// Date/Time types
		{"DATE", "CODE_DATE", false},
		{"DATETIME", "CODE_TIMESTAMP", false},
		{"TIMESTAMP", "CODE_TIMESTAMP", false},

		// Semi-structured types
		{"JSON", "CODE_JSON", false},
		{"JSONB", "CODE_JSON", false},
		{"ARRAY", "CODE_ARRAY", false},
		{"MAP", "CODE_MAP", false},
		{"STRUCT", "CODE_STRUCT", false},

		// Unsupported types (aggregate-only types)
		{"HLL", "", true},
		{"BITMAP", "", true},
		{"PERCENTILE", "", true},
		{"UNKNOWN_TYPE", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.dbType, func(t *testing.T) {
			result, err := c.databaseTypeToRuntimeType(tt.dbType)
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Contains(t, result.Code.String(), tt.expected)
		})
	}
}
