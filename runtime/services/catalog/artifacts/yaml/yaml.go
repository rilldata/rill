// Package yaml reads and writes artifacts that exactly mirror the internal representation
package yaml

import (
	"context"
	"errors"
	"path"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
)

type artifact struct{}

var NotSupported = errors.New("yaml only supported for sources and dashboards")

func init() {
	artifacts.Register(".yaml", &artifact{})
}

func (r *artifact) DeSerialise(ctx context.Context, filePath string, blob string) (*api.CatalogObject, error) {
	ext := fileutil.FullExt(filePath)
	fileName := path.Base(filePath)
	dir := strings.Trim(path.Dir(filePath), "./")
	name := strings.TrimSuffix(fileName, ext)

	switch dir {
	case "sources":
		source := &Source{}
		err := yaml.Unmarshal([]byte(blob), &source)
		if err != nil {
			return nil, err
		}
		return fromSourceArtifact(name, fileName, source)
	case "dashboards":
		metrics := &MetricsView{}
		err := yaml.Unmarshal([]byte(blob), &metrics)
		if err != nil {
			return nil, err
		}
		return fromMetricsViewArtifact(name, fileName, metrics)
	}

	return nil, NotSupported
}

func (r *artifact) Serialise(ctx context.Context, catalogObject *api.CatalogObject) (string, error) {
	switch catalogObject.Type {
	case api.CatalogObject_TYPE_SOURCE:
		source, err := toSourceArtifact(catalogObject)
		if err != nil {
			return "", err
		}
		out, err := yaml.Marshal(source)
		if err != nil {
			return "", err
		}
		return string(out), nil
	case api.CatalogObject_TYPE_METRICS_VIEW:
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
