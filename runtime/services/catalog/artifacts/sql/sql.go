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

func (r *artifact) DeSerialise(ctx context.Context, filePath, blob string, materializeDefault bool) (*drivers.CatalogEntry, error) {
	name := fileutil.Stem(filePath)
	// extract materialize option before sanitizing query as it will remove that comment
	materialize := parseMaterializationInfo(blob)
	if materialize == MaterializeInvalid {
		return nil, errors.New("invalid materialize type")
	}
	if materialize == MaterializeInferred {
		if materializeDefault {
			materialize = MaterializeTrue
		} else {
			materialize = MaterializeFalse
		}
	}
	sanitizedSQL := sanitizeQuery(blob)
	return &drivers.CatalogEntry{
		Type: drivers.ObjectTypeModel,
		Object: &runtimev1.Model{
			Name:        name,
			Sql:         sanitizedSQL,
			Dialect:     runtimev1.Model_DIALECT_DUCKDB,
			Materialize: materialize.Materialize(),
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
	MaterializedRegex     = regexp.MustCompile(`(?m)^--[ \t]*@materialize[ \t]?:[ \t]*([a-zA-Z]*)\s+`)
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

// MaterializationInfo Materialization values for models, specified using @materialize: tag in the comment
type MaterializationInfo int64

const (
	// MaterializeUnspecified When tag is not specified
	MaterializeUnspecified MaterializationInfo = iota
	// MaterializeTrue When tag is specified as true
	MaterializeTrue
	// MaterializeFalse When tag is specified as false
	MaterializeFalse
	// MaterializeInferred When it is not specified by the user, but we infer it and set this value
	MaterializeInferred
	// MaterializeInvalid When tag is specified but value is either empty or invalid
	MaterializeInvalid
)

func (m MaterializationInfo) Materialize() bool {
	switch m {
	case MaterializeTrue:
		return true
	case MaterializeInferred:
		return true
	default:
		return false
	}
}

func parseMaterializationInfo(query string) MaterializationInfo {
	matched := MaterializedRegex.FindStringSubmatch(query)
	if len(matched) == 0 {
		return MaterializeUnspecified
	}
	switch strings.ToLower(matched[1]) {
	case "true":
		return MaterializeTrue
	case "false":
		return MaterializeFalse
	case "inferred":
		return MaterializeInferred
	default:
		return MaterializeInvalid
	}
}
