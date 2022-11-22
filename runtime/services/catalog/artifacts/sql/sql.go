package sql

import (
	"context"
	"errors"
	"path"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
)

/**
 * this package contains code to map an sql file to a catalog object
 */

type artifact struct{}

var NotSupported = errors.New("only model supported for sql")

func init() {
	artifacts.Register(".sql", &artifact{})
}

func (r *artifact) DeSerialise(ctx context.Context, filePath string, blob string) (*runtimev1.CatalogObject, error) {
	ext := fileutil.FullExt(filePath)
	fileName := path.Base(filePath)
	name := strings.TrimSuffix(fileName, ext)
	return &runtimev1.CatalogObject{
		Type: runtimev1.CatalogObject_TYPE_MODEL,
		Model: &runtimev1.Model{
			Name:    name,
			Sql:     blob,
			Dialect: runtimev1.Model_DIALECT_DUCKDB,
		},
		Name: name,
		Path: filePath,
	}, nil
}

func (r *artifact) Serialise(ctx context.Context, catalogObject *runtimev1.CatalogObject) (string, error) {
	if catalogObject.Type != runtimev1.CatalogObject_TYPE_MODEL {
		return "", NotSupported
	}
	return catalogObject.Model.Sql, nil
}
