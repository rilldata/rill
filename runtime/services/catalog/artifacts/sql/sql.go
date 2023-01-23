package sql

import (
	"context"
	"errors"
	"regexp"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
)

/**
 * this package contains code to map an sql file to a catalog object
 */

type artifact struct{}

var ErrNotSupported = errors.New("only model supported for sql")

func init() {
	artifacts.Register(".sql", &artifact{})
}

func (r *artifact) DeSerialise(ctx context.Context, filePath, blob string) (*drivers.CatalogEntry, error) {
	name := fileutil.Stem(filePath)
	// extract materialize option before sanitizing query as it will remove that comment
	materialize := parseMaterializationInfo(blob)
	sanitizedSql := sanitizeQuery(blob)
	return &drivers.CatalogEntry{
		Type: drivers.ObjectTypeModel,
		Object: &runtimev1.Model{
			Name:        name,
			Sql:         sanitizedSql,
			Dialect:     runtimev1.Model_DIALECT_DUCKDB,
			Materialize: materialize,
		},
		Name: name,
		Path: filePath,
	}, nil
}

func (r *artifact) Serialise(ctx context.Context, catalogObject *drivers.CatalogEntry) (string, error) {
	if catalogObject.Type != drivers.ObjectTypeModel {
		return "", ErrNotSupported
	}
	return catalogObject.GetModel().Sql, nil
}

var (
	QueryCommentRegex     = regexp.MustCompile(`(?m)--.*$`)
	MultipleSpacesRegex   = regexp.MustCompile(`\s\s+`)
	SpacesAfterCommaRegex = regexp.MustCompile(`,\s+`)
	MaterializedRegex     = regexp.MustCompile(`--\s*@materialize[ |\t]?:[ |\t]*([a-zA-Z]*)\s+`)
)

// TODO: use this while extracting source names to get case insensitive dag
// TODO: should this be used to store the sql in catalog?
func sanitizeQuery(query string) string {
	// remove all comments
	query = QueryCommentRegex.ReplaceAllString(query, " ")
	// new line => space
	query = strings.ReplaceAll(query, "\n", " ")
	// multiple spaces => single space
	query = MultipleSpacesRegex.ReplaceAllString(query, " ")
	// remove all spaces after a comma
	query = SpacesAfterCommaRegex.ReplaceAllString(query, ",")
	query = strings.ReplaceAll(query, ";", "")
	return strings.TrimSpace(query)
}

func parseMaterializationInfo(query string) runtimev1.Model_Materialize {
	matched := MaterializedRegex.FindStringSubmatch(query)
	if len(matched) == 0 {
		return runtimev1.Model_MATERIALIZE_UNSPECIFIED
	}
	switch strings.ToLower(matched[1]) {
	case "true":
		return runtimev1.Model_MATERIALIZE_TRUE
	case "false":
		return runtimev1.Model_MATERIALIZE_FALSE
	case "inferred":
		return runtimev1.Model_MATERIALIZE_INFERRED
	default:
		return runtimev1.Model_MATERIALIZE_INVALID
	}
}
