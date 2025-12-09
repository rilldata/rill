package sqlstring

import (
	"testing"
	"time"
)

func TestToSQLLiteral(t *testing.T) {
	fixedTime := time.Date(2024, 3, 15, 10, 30, 45, 123456789, time.UTC)

	tests := []struct {
		name string
		val  any
		want string
	}{
		// Nil
		{"nil", nil, "NULL"},

		// Strings
		{"string", "hello", "'hello'"},
		{"string with quote", "it's", "'it''s'"},
		{"string multiple quotes", "a''b", "'a''''b'"},
		{"empty string", "", "''"},

		// Integers
		{"int", 42, "42"},
		{"int negative", -42, "-42"},
		{"int8", int8(127), "127"},
		{"int16", int16(-32000), "-32000"},
		{"int32", int32(2147483647), "2147483647"},
		{"int64", int64(9223372036854775807), "9223372036854775807"},

		// Unsigned integers
		{"uint", uint(42), "42"},
		{"uint8", uint8(255), "255"},
		{"uint16", uint16(65535), "65535"},
		{"uint32", uint32(4294967295), "4294967295"},
		{"uint64", uint64(18446744073709551615), "18446744073709551615"},

		// Floats
		{"float32", float32(3.14), "3.14"},
		{"float32 whole", float32(42), "42"},
		{"float64", 3.14159265359, "3.14159265359"},
		{"float64 scientific", 1.23e10, "12300000000"},
		{"float64 negative", -0.001, "-0.001"},

		// Booleans
		{"bool true", true, "TRUE"},
		{"bool false", false, "FALSE"},

		// Time
		{"time.Time", fixedTime, "'2024-03-15T10:30:45.123456789Z'"},
		{"*time.Time", &fixedTime, "'2024-03-15T10:30:45.123456789Z'"},
		{"*time.Time nil", (*time.Time)(nil), "NULL"},

		// Pointers to primitives
		{"*string", ptr("hello"), "'hello'"},
		{"*string nil", (*string)(nil), "NULL"},
		{"*int", ptr(42), "42"},
		{"*int nil", (*int)(nil), "NULL"},
		{"*bool", ptr(true), "TRUE"},
		{"**int", ptr(ptr(42)), "42"},
		{"**int outer nil", (**int)(nil), "NULL"},

		// Slices
		{"[]int", []int{1, 2, 3}, "(1, 2, 3)"},
		{"[]int empty", []int{}, "(NULL)"},
		{"[]string", []string{"a", "b"}, "('a', 'b')"},
		{"[]string with quotes", []string{"it's", "ok"}, "('it''s', 'ok')"},
		{"[]bool", []bool{true, false}, "(TRUE, FALSE)"},
		{"[]any mixed", []any{1, "two", true}, "(1, 'two', TRUE)"},

		// Byte slices (hex encoding)
		{"[]byte", []byte{0xde, 0xad, 0xbe, 0xef}, "X'deadbeef'"},
		{"[]byte empty", []byte{}, "X''"},

		// Arrays
		{"[3]int", [3]int{1, 2, 3}, "(1, 2, 3)"},
		{"[0]int", [0]int{}, "(NULL)"},

		// Nested slices
		{"[][]int", [][]int{{1, 2}, {3, 4}}, "((1, 2), (3, 4))"},

		// Fallback to JSON
		{"struct", struct{ X int }{X: 42}, "'{\"X\":42}'"},
		{"map", map[string]int{"a": 1}, "'{\"a\":1}'"},
		{"struct with quotes", struct{ S string }{S: "it's"}, "'{\"S\":\"it''s\"}'"},
		{"unsupported type", make(chan int), "'<json error: json: unsupported type: chan int>'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToLiteral(tt.val)
			if got != tt.want {
				t.Errorf("ToSQLLiteral(%v) = %q, want %q", tt.val, got, tt.want)
			}
		})
	}
}

func ptr[T any](v T) *T { return &v }
