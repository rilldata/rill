// Package yaml reads and writes artifacts that exactly mirror the internal representation
package yaml

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
	"gopkg.in/yaml.v3"
)

type artifact struct{}

var ErrNotSupported = errors.New("yaml only supported for sources and dashboards")

func init() {
	artifacts.Register(".yaml", &artifact{})
}

func (r *artifact) DeSerialise(ctx context.Context, filePath, blob string) (*drivers.CatalogEntry, error) {
	dir := filepath.Base(filepath.Dir(filePath))
	switch dir {
	case "sources":
		source := &Source{}
		err := yaml.Unmarshal([]byte(blob), &source)
		if err != nil {
			return nil, err
		}
		return fromSourceArtifact(source, filePath)
	case "dashboards":
		metrics := &MetricsView{}
		err := yaml.Unmarshal([]byte(blob), &metrics)
		if err != nil {
			return nil, err
		}
		return fromMetricsViewArtifact(metrics, filePath)
	}

	return nil, ErrNotSupported
}

func (r *artifact) Serialise(ctx context.Context, catalogObject *drivers.CatalogEntry) (string, error) {
	switch catalogObject.Type {
	case drivers.ObjectTypeSource:
		source, err := toSourceArtifact(catalogObject)
		if err != nil {
			return "", err
		}
		out, err := yaml.Marshal(source)
		if err != nil {
			return "", err
		}
		return string(out), nil
	case drivers.ObjectTypeMetricsView:
		metrics, err := toMetricsViewArtifact(catalogObject)
		if err != nil {
			return "", err
		}
		out, err := yaml.Marshal(metrics)
		if err != nil {
			return "", err
		}
		return string(out), nil
	}

	return "", ErrNotSupported
}
