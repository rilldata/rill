package yaml

import (
	"context"

	"github.com/go-yaml/yaml"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog/artifacts"
)

type artifact struct{}

func init() {
	artifacts.Register(".yaml", &artifact{})
}

func (r *artifact) DeSerialise(ctx context.Context, blob string) (*drivers.CatalogObject, error) {
	var artifactObject Artifact
	err := yaml.Unmarshal([]byte(blob), &artifactObject)
	if err != nil {
		return nil, err
	}
	return fromArtifact(&artifactObject)
}

func (r *artifact) Serialise(ctx context.Context, catalogObject *drivers.CatalogObject) (string, error) {
	artifact, err := toArtifact(catalogObject)
	if err != nil {
		return "", err
	}
	out, err := yaml.Marshal(artifact)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
