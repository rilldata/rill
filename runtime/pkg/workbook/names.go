package workbook

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Excel limits sheet names to 31 characters.
const maxSheetNameLen = 31

// invalidSheetNameChars are the characters Excel does not allow in sheet names.
const invalidSheetNameChars = `\/?*[]:`

// namer allocates unique sheet and table names within one workbook.
type namer struct {
	// Sheet names in use, lower-cased: Excel compares sheet names case-insensitively.
	sheets map[string]bool
	tables map[string]bool
}

func newNamer() *namer {
	return &namer{sheets: map[string]bool{}, tables: map[string]bool{}}
}

// sheetName returns a valid, unused sheet name derived from base.
func (n *namer) sheetName(base string) string {
	name := strings.Map(func(r rune) rune {
		if strings.ContainsRune(invalidSheetNameChars, r) || unicode.IsControl(r) {
			return -1
		}
		return r
	}, base)
	// A sheet name can't start or end with an apostrophe, including after it is cut to length.
	name = trimSheetName(truncateRunes(trimSheetName(name), maxSheetNameLen))
	if name == "" {
		name = "Sheet"
	}
	// Excel reserves "History" for its change tracking.
	if strings.EqualFold(name, "History") {
		name = "History sheet"
	}

	candidate := name
	for i := 2; n.sheets[strings.ToLower(candidate)]; i++ {
		suffix := fmt.Sprintf(" (%d)", i)
		candidate = trimSheetName(truncateRunes(name, maxSheetNameLen-len(suffix))) + suffix
	}
	n.sheets[strings.ToLower(candidate)] = true
	return candidate
}

// tableName returns a valid, unused Excel table name derived from base.
// Table names may only hold letters, digits and underscores, so every other run of characters becomes one underscore.
// The "Table_" prefix keeps the name from starting with a digit or looking like a cell reference.
func (n *namer) tableName(base string) string {
	var b strings.Builder
	b.WriteString("Table")
	underscore := true
	for _, r := range base {
		if r < utf8.RuneSelf && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
			if underscore {
				b.WriteByte('_')
			}
			b.WriteRune(r)
			underscore = false
		} else {
			underscore = true
		}
	}
	name := truncateRunes(b.String(), 200)

	candidate := name
	for i := 2; n.tables[strings.ToLower(candidate)]; i++ {
		candidate = fmt.Sprintf("%s_%d", name, i)
	}
	n.tables[strings.ToLower(candidate)] = true
	return candidate
}

// quoteSheet quotes a sheet name for use in a formula, e.g. 'Data - Spend'!$A$1.
func quoteSheet(name string) string {
	return "'" + strings.ReplaceAll(name, "'", "''") + "'"
}

// trimSheetName strips the spaces and apostrophes a sheet name can't start or end with.
func trimSheetName(name string) string {
	for {
		trimmed := strings.TrimSpace(strings.Trim(name, "'"))
		if trimmed == name {
			return name
		}
		name = trimmed
	}
}

func truncateRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return strings.TrimSpace(string([]rune(s)[:n]))
}
