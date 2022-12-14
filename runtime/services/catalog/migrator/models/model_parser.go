package models

import (
	"regexp"
)

/**
 * Parses a model query but looks specifically for source table names
 */

var tableNameRegex = regexp.MustCompile(`(?im)(?:from|join)\s+([a-zA-z0-9_.]+|"[a-zA-z0-9\\.\\-_/:\s]+")`)

func ExtractTableNames(query string) []string {
	subMatches := tableNameRegex.FindAllStringSubmatch(query, -1)
	var tableNames []string
	dedupeMap := make(map[string]bool)

	for _, subMatch := range subMatches {
		if len(subMatch) >= 2 && !dedupeMap[subMatch[1]] {
			tableNames = append(tableNames, subMatch[1])
			dedupeMap[subMatch[1]] = true
		}
	}
	return tableNames
}
