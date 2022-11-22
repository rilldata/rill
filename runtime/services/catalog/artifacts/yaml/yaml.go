// Package yaml reads and writes artifacts that exactly mirror the internal representation
package yaml

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/go-yaml/yaml"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
)

type artifact struct{}

var NotSupported = errors.New("yaml only supported for sources and dashboards")

func init() {
	artifacts.Register(".yaml", &artifact{})
}

func (r *artifact) DeSerialise(ctx context.Context, filePath string, blob string) (*runtimev1.CatalogObject, error) {
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

	return nil, NotSupported
}

func (r *artifact) Serialise(ctx context.Context, catalogObject *runtimev1.CatalogObject) (string, error) {
	switch catalogObject.Type {
	case runtimev1.CatalogObject_TYPE_SOURCE:
		source, err := toSourceArtifact(catalogObject)
		if err != nil {
			return "", err
		}
		out, err := yaml.Marshal(source)
		if err != nil {
			return "", err
		}
		return string(out), nil
	case runtimev1.CatalogObject_TYPE_METRICS_VIEW:
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

	return "", NotSupported
}
