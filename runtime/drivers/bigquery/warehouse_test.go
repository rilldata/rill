package bigquery

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelectQueryRegex(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		shouldMatch   bool
		expectedTable string
	}{
		// Valid cases - basic table names
		{
			name:          "simple table name",
			query:         "SELECT * FROM table",
			shouldMatch:   true,
			expectedTable: "table",
		},
		{
			name:          "table name with backticks",
			query:         "SELECT * FROM `table`",
			shouldMatch:   true,
			expectedTable: "`table`",
		},
		{
			name:          "dataset.table format",
			query:         "SELECT * FROM dataset.table",
			shouldMatch:   true,
			expectedTable: "dataset.table",
		},
		{
			name:          "dataset.table with backticks",
			query:         "SELECT * FROM `dataset.table`",
			shouldMatch:   true,
			expectedTable: "`dataset.table`",
		},
		{
			name:          "project.dataset.table format",
			query:         "SELECT * FROM project.dataset.table",
			shouldMatch:   true,
			expectedTable: "project.dataset.table",
		},
		{
			name:          "project.dataset.table with backticks",
			query:         "SELECT * FROM `project.dataset.table`",
			shouldMatch:   true,
			expectedTable: "`project.dataset.table`",
		},
		{
			name:          "table name with underscores and hyphens",
			query:         "SELECT * FROM `my-table_name`",
			shouldMatch:   true,
			expectedTable: "`my-table_name`",
		},
		{
			name:          "table name with numbers",
			query:         "SELECT * FROM table123",
			shouldMatch:   true,
			expectedTable: "table123",
		},

		// Valid cases - with whitespace
		{
			name:          "with leading whitespace",
			query:         "  SELECT * FROM table",
			shouldMatch:   true,
			expectedTable: "table",
		},
		{
			name:          "with trailing whitespace",
			query:         "SELECT * FROM table  ",
			shouldMatch:   true,
			expectedTable: "table",
		},
		{
			name:          "with multiple spaces",
			query:         "SELECT   *   FROM   table",
			shouldMatch:   true,
			expectedTable: "table",
		},
		{
			name:          "with newlines",
			query:         "SELECT * FROM\n  table",
			shouldMatch:   true,
			expectedTable: "table",
		},
		{
			name:          "with semicolon",
			query:         "SELECT * FROM table;",
			shouldMatch:   true,
			expectedTable: "table",
		},
		{
			name:          "with semicolon and whitespace",
			query:         "SELECT * FROM table ; ",
			shouldMatch:   true,
			expectedTable: "table",
		},

		// Valid cases - with comments
		{
			name:          "with leading single-line comment (--)",
			query:         "-- This is a comment\nSELECT * FROM table",
			shouldMatch:   true,
			expectedTable: "table",
		},
		{
			name:          "with leading hash comment",
			query:         "# This is a comment\nSELECT * FROM table",
			shouldMatch:   true,
			expectedTable: "table",
		},
		{
			name:          "with leading multi-line comment",
			query:         "/* This is a comment */\nSELECT * FROM table",
			shouldMatch:   true,
			expectedTable: "table",
		},
		{
			name:          "with trailing single-line comment",
			query:         "SELECT * FROM table -- comment",
			shouldMatch:   true,
			expectedTable: "table",
		},
		{
			name:          "with trailing hash comment",
			query:         "SELECT * FROM table # comment",
			shouldMatch:   true,
			expectedTable: "table",
		},
		{
			name:          "with trailing multi-line comment",
			query:         "SELECT * FROM table /* comment */",
			shouldMatch:   true,
			expectedTable: "table",
		},
		{
			name:          "with multiple leading comments",
			query:         "-- First comment\n# Second comment\n/* Third comment */\nSELECT * FROM table",
			shouldMatch:   true,
			expectedTable: "table",
		},
		{
			name:          "with comments and semicolon",
			query:         "SELECT * FROM table; -- comment",
			shouldMatch:   true,
			expectedTable: "table",
		},

		// Invalid cases
		{
			name:        "SELECT with specific columns",
			query:       "SELECT col1, col2 FROM table",
			shouldMatch: false,
		},
		{
			name:        "SELECT with WHERE clause",
			query:       "SELECT * FROM table WHERE id = 1",
			shouldMatch: false,
		},
		{
			name:        "missing asterisk",
			query:       "SELECT col FROM table",
			shouldMatch: false,
		},
		{
			name:        "partial match in larger query",
			query:       "SELECT * FROM table WHERE 1=1",
			shouldMatch: false,
		},
		{
			name:        "SELECT * FROM with LIMIT",
			query:       "SELECT * FROM table LIMIT 10",
			shouldMatch: false,
		},
		{
			name:        "SELECT * FROM with ORDER BY",
			query:       "SELECT * FROM table ORDER BY id",
			shouldMatch: false,
		},
		{
			name:        "empty string",
			query:       "",
			shouldMatch: false,
		},
		{
			name:        "just whitespace",
			query:       "   ",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := selectQueryRegex.FindStringSubmatch(tt.query)
			if tt.shouldMatch {
				require.NotNil(t, match, "Expected query to match: %q", tt.query)
				require.Len(t, match, 2, "Expected 2 capture groups (full match + table name)")
				assert.Equal(t, tt.expectedTable, match[1], "Table name should match")
			} else {
				assert.Nil(t, match, "Expected query NOT to match: %q", tt.query)
			}
		})
	}
}
