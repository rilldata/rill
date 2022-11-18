package sql

import (
	"context"
	"errors"
	"path"
	"strings"

	"github.com/rilldata/rill/runtime/api"
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

func (r *artifact) DeSerialise(ctx context.Context, filePath string, blob string) (*api.CatalogObject, error) {
	ext := fileutil.FullExt(filePath)
	fileName := path.Base(filePath)
	name := strings.TrimSuffix(fileName, ext)
	return &api.CatalogObject{
		Type: api.CatalogObject_TYPE_MODEL,
		Model: &api.Model{
			Name:    name,
			Sql:     blob,
			Dialect: api.Model_DIALECT_DUCKDB,
		},
		Name: name,
		Path: filePath,
	}, nil
}

func (r *artifact) Serialise(ctx context.Context, catalogObject *api.CatalogObject) (string, error) {
	if catalogObject.Type != api.CatalogObject_TYPE_MODEL {
		return "", NotSupported
	}
	return catalogObject.Model.Sql, nil
}
