package duckdbsql

import "fmt"

// EvaluateBool uses DuckDB to evaluate the given expression as a boolean value.
// It may succeed on some non-bool expressions, such as integer expressions, because they are castable to a bool.
func EvaluateBool(expr string) (bool, error) {
	res, err := queryAny(fmt.Sprintf("SELECT (%s)::BOOLEAN", expr))
	if err != nil {
		return false, err
	}
	b, ok := res.(bool)
	if !ok {
		return false, fmt.Errorf("internal: expected bool, got %v", res)
	}
	return b, err
}
