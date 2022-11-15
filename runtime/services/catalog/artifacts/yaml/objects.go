package yaml

import (
	"fmt"

	"github.com/jinzhu/copier"
	"github.com/rilldata/rill/runtime/api"
	"google.golang.org/protobuf/types/known/structpb"
)

/**
 * This file contains the mapping from CatalogObject to Yaml files
 */

const Version = "0.0.1"

type Source struct {
	Version string
	Type    string
	URI     string
	Region  string `yaml:"region,omitempty"`
}

type MetricsView struct {
	Version          string
	DisplayName      string `yaml:"display_name"`
	Description      string
	From             string
	TimeDimension    string `yaml:"time_dimension"`
	TimeGrains       []string
	DefaultTimeGrain string `yaml:"default_timegrain"`
	Dimensions       []*Dimension
	Measures         []*Measure
}

type Measure struct {
	Label       string
	Expression  string
	Description string
	Format      string `yaml:"format_preset"`
}

type Dimension struct {
	Label       string
	Property    string `copier:"Name"`
	Description string
}

func toSourceArtifact(catalog *api.CatalogObject) (*Source, error) {
	source := &Source{
		Version: Version,
		Type:    catalog.Source.Connector,
	}

	props := catalog.Source.Properties.AsMap()
	uri, ok := props["path"].(string)
	if ok {
		source.URI = uri
	}
	region, ok := props["region"].(string)
	if ok {
		source.Region = region
	}

	return source, nil
}

func toMetricsViewArtifact(catalog *api.CatalogObject) (*MetricsView, error) {
	metricsArtifact := &MetricsView{}
	err := copier.Copy(metricsArtifact, catalog.MetricsView)
	if err != nil {
		return nil, err
	}

	metricsArtifact.Version = Version
	return metricsArtifact, nil
}

func fromSourceArtifact(name string, path string, source *Source) (*api.CatalogObject, error) {
	props := map[string]interface{}{
		"path": source.URI,
	}
	if source.Region != "" {
		props["region"] = source.Region
	}
	propsPB, err := structpb.NewStruct(props)
	if err != nil {
		return nil, err
	}

	return &api.CatalogObject{
		Name: name,
		Type: api.CatalogObject_TYPE_SOURCE,
		Path: path,
		Source: &api.Source{
			Name:       name,
			Connector:  source.Type,
			Properties: propsPB,
		},
	}, nil
}

func fromMetricsViewArtifact(name string, path string, metrics *MetricsView) (*api.CatalogObject, error) {
	apiMetrics := &api.MetricsView{}
	err := copier.Copy(apiMetrics, metrics)
	if err != nil {
		return nil, err
	}
	// this is needed since measure names are not given by the user
	for i, measure := range apiMetrics.Measures {
		measure.Name = fmt.Sprintf("measure_%d", i)
	}

	apiMetrics.Name = name
	return &api.CatalogObject{
		Name:        name,
		Type:        api.CatalogObject_TYPE_METRICS_VIEW,
		Path:        path,
		MetricsView: apiMetrics,
	}, nil
}
