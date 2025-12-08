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
				DSN: "starrocks://user:pass@host:9030/db",
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
				DSN:  "starrocks://user:pass@host:9030/db",
				Host: "localhost",
			},
			wantErr: false,
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

func TestConvertDSN(t *testing.T) {
	c := &connection{
		configProp: &ConfigProperties{},
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "mysql format passthrough",
			input:    "user:pass@tcp(host:9030)/db",
			expected: "user:pass@tcp(host:9030)/?timeout=30s&readTimeout=300s&writeTimeout=30s&parseTime=true",
		},
		{
			name:     "starrocks url format",
			input:    "starrocks://user:pass@host:9030/db",
			expected: "user:pass@tcp(host:9030)/?timeout=30s&readTimeout=300s&writeTimeout=30s&parseTime=true",
		},
		{
			name:     "starrocks url without user",
			input:    "starrocks://host:9030/db",
			expected: "tcp(host:9030)/?timeout=30s&readTimeout=300s&writeTimeout=30s&parseTime=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := c.convertDSN(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.expected, result)
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

func TestEscapeReservedKeyword(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "reserved keyword range",
			input:    "range",
			expected: "valRange",
		},
		{
			name:     "reserved keyword values",
			input:    "values",
			expected: "vals",
		},
		{
			name:     "reserved keyword RANGE uppercase",
			input:    "RANGE",
			expected: "valRange",
		},
		{
			name:     "non-reserved word",
			input:    "column_name",
			expected: "column_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeReservedKeyword(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestDatabaseTypeToRuntimeType(t *testing.T) {
	c := &connection{}

	tests := []struct {
		dbType   string
		expected string
	}{
		{"BOOLEAN", "CODE_BOOL"},
		{"INT", "CODE_INT32"},
		{"BIGINT", "CODE_INT64"},
		{"DOUBLE", "CODE_FLOAT64"},
		{"VARCHAR(255)", "CODE_STRING"},
		{"DATETIME", "CODE_TIMESTAMP"},
		{"DATE", "CODE_DATE"},
		{"JSON", "CODE_JSON"},
		{"NULLABLE(INT)", "CODE_INT32"},
		{"UNKNOWN_TYPE", "CODE_STRING"}, // default fallback
	}

	for _, tt := range tests {
		t.Run(tt.dbType, func(t *testing.T) {
			result := c.databaseTypeToRuntimeType(tt.dbType)
			require.Contains(t, result.Code.String(), tt.expected)
		})
	}
}
