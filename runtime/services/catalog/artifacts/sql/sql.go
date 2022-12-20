package sql

import (
	"context"
	"errors"

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
	return &drivers.CatalogEntry{
		Type: drivers.ObjectTypeModel,
		Object: &runtimev1.Model{
			Name:    name,
			Sql:     blob,
			Dialect: runtimev1.Model_DIALECT_DUCKDB,
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
