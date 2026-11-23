// Package xlsxutil holds helpers shared by the writers of Excel files:
// the XLSX exports of driverutil and the canvas workbook.
package xlsxutil

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Column widths, in characters.
const (
	MinColumnWidth = 10
	MaxColumnWidth = 60
)

// UniqueNames returns the names made unique and non-empty, which an Excel table's column names must be.
// Excel compares them ignoring case.
// A name repeated after its first use gets the first free suffix, like "Country (2)",
// so the names that are already unique, like a real "Country (2)", are kept.
func UniqueNames(names []string) []string {
	res := make([]string, len(names))
	used := make(map[string]bool, len(names))
	for i, name := range names {
		if name == "" {
			name = fmt.Sprintf("Column %d", i+1)
		}
		res[i] = name
		used[strings.ToLower(name)] = true
	}
	seen := make(map[string]bool, len(names))
	for i, name := range res {
		if key := strings.ToLower(name); !seen[key] {
			seen[key] = true
			continue
		}
		for n := 2; ; n++ {
			candidate := fmt.Sprintf("%s (%d)", name, n)
			if key := strings.ToLower(candidate); !used[key] {
				used[key] = true
				seen[key] = true
				res[i] = candidate
				break
			}
		}
	}
	return res
}

// CellWidth estimates the width of a cell value in characters, for sizing its column.
func CellWidth(v any) int {
	var n float64
	switch v := v.(type) {
	case nil:
		return 0
	case string:
		return utf8.RuneCountInString(v)
	case bool:
		return 5
	case time.Time:
		return 19
	case int64:
		n = float64(v)
	case uint64:
		n = float64(v)
	case float64:
		n = v
	default:
		return 10
	}
	// Digits, two decimals, thousands separators and a currency symbol.
	digits := len(strconv.FormatFloat(n, 'f', 2, 64))
	return digits + digits/3 + 1
}

// ColumnWidth returns the width of a column whose widest value is chars characters wide:
// with some padding, and within MinColumnWidth and MaxColumnWidth.
func ColumnWidth(chars int) float64 {
	return math.Min(math.Max(float64(chars)+2, MinColumnWidth), MaxColumnWidth)
}
