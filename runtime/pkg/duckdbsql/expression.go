package duckdbsql

import (
	"errors"
	"fmt"
)

// EvaluateBool uses DuckDB to evaluate the given expression as a boolean value.
func EvaluateBool(expr string) (bool, error) {
	// NOTE: Doing type check in SQL instead of in Go to protect against expressions that return huge values.
	res, isBool, err := queryBoolPair(fmt.Sprintf("SELECT v::BOOLEAN AS b, TYPEOF(v) = 'BOOLEAN' AS is_bool FROM (SELECT (%s) AS v)", expr))
	if err != nil {
		return false, err
	}
	if !isBool {
		return false, errors.New("expression did not evaluate to a bool")
	}
	return res, nil
}

// queryBoolPair runs a DuckDB query and returns the result as a pair of bools.
func queryBoolPair(qry string, args ...any) (bool, bool, error) {
	rows, err := query(qry, args...)
	if err != nil {
		return false, false, err
	}
	defer func() { _ = rows.Close() }()

	var a, b bool
	if rows.Next() {
		err := rows.Scan(&a, &b)
		if err != nil {
			return false, false, err
		}
	}

	return a, b, nil
}
