package yaml

import (
	"fmt"

	"github.com/jinzhu/copier"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/structpb"
)

/**
 * This file contains the mapping from CatalogObject to Yaml files
 */

const Version = "0.0.1"

type Artifact struct {
	Version    string
	Type       string
	Definition any
}

type Source struct {
	Name       string
	Connector  string
	Properties map[string]any
}

type Model struct {
	Name    string
	Sql     string
	Dialect string
}

const (
	ModelDialectDuckDB string = "duckdb"
)

type MetricsView struct {
	Name          string
	From          string
	TimeDimension string
	TimeGrains    []string
	Dimensions    []*Dimension
	Measures      []*Measure
}

type Measure struct {
	Label       string
	Expression  string
	Description string
	Format      string
}

type Dimension struct {
	Label       string
	Property    string `copier:"Name"`
	Description string
}

func toArtifact(catalog *api.CatalogObject) (*Artifact, error) {
	artifact := Artifact{
		Version: Version,
	}

	var err error
	switch catalog.Type.(type) {
	case *api.CatalogObject_Source:
		artifact.Definition, err = toSourceArtifact(catalog)
		artifact.Type = drivers.CatalogObjectTypeSource
	case *api.CatalogObject_Model:
		artifact.Definition, err = toModelArtifact(catalog)
		artifact.Type = drivers.CatalogObjectTypeModel
	case *api.CatalogObject_MetricsView:
		artifact.Definition, err = toMetricsViewArtifact(catalog)
		artifact.Type = drivers.CatalogObjectTypeMetricsView
	}
	if err != nil {
		return nil, err
	}

	return &artifact, nil
}

func toSourceArtifact(catalog *api.CatalogObject) (*Source, error) {
	source, ok := catalog.Type.(*api.CatalogObject_Source)
	if !ok {
		return nil, fmt.Errorf("failed to parse source")
	}

	return &Source{
		Name:       source.Source.Name,
		Connector:  source.Source.Connector,
		Properties: source.Source.Properties.AsMap(),
	}, nil
}

func toModelArtifact(catalog *api.CatalogObject) (*Model, error) {
	model, ok := catalog.Type.(*api.CatalogObject_Model)
	if !ok {
		return nil, fmt.Errorf("failed to parse model")
	}

	modelArtifact := &Model{
		Name: model.Model.Name,
		Sql:  model.Model.Sql,
	}

	switch model.Model.Dialect {
	case api.Model_DuckDB:
		modelArtifact.Dialect = ModelDialectDuckDB
	}

	return modelArtifact, nil
}

func toMetricsViewArtifact(catalog *api.CatalogObject) (*MetricsView, error) {
	metricsView, ok := catalog.Type.(*api.CatalogObject_MetricsView)
	if !ok {
		return nil, fmt.Errorf("failed to parse metrics view")
	}

	metricsArtifact := &MetricsView{}
	err := copier.Copy(metricsArtifact, &metricsView.MetricsView)
	if err != nil {
		return nil, err
	}

	return metricsArtifact, nil
}

func fromArtifact(artifact *Artifact) (*api.CatalogObject, error) {
	catalog := &api.CatalogObject{}

	var err error
	switch artifact.Type {
	case drivers.CatalogObjectTypeSource:
		err = fromSourceArtifact(artifact, catalog)
	case drivers.CatalogObjectTypeModel:
		err = fromModelArtifact(artifact, catalog)
	case drivers.CatalogObjectTypeMetricsView:
		err = fromMetricsViewArtifact(artifact, catalog)
	}

	if err != nil {
		return nil, err
	}
	return catalog, nil
}

func fromSourceArtifact(artifact *Artifact, catalog *api.CatalogObject) error {
	var sourceArtifact Source
	err := mapstructure.Decode(artifact.Definition, &sourceArtifact)
	if err != nil {
		return err
	}

	propsPB, err := structpb.NewStruct(sourceArtifact.Properties)
	if err != nil {
		return err
	}

	catalog.Name = sourceArtifact.Name
	catalog.Type = &api.CatalogObject_Source{
		Source: &api.Source{
			Name:       catalog.Name,
			Connector:  sourceArtifact.Connector,
			Properties: propsPB,
		},
	}

	return nil
}

func fromModelArtifact(artifact *Artifact, catalog *api.CatalogObject) error {
	var modelArtifact Model
	err := mapstructure.Decode(artifact.Definition, &modelArtifact)
	if err != nil {
		return err
	}

	catalog.Name = modelArtifact.Name
	model := &api.Model{
		Name: modelArtifact.Name,
		Sql:  modelArtifact.Sql,
	}
	catalog.Type = &api.CatalogObject_Model{
		Model: model,
	}
	switch modelArtifact.Dialect {
	case ModelDialectDuckDB:
		model.Dialect = api.Model_DuckDB
	}

	return nil
}

func fromMetricsViewArtifact(artifact *Artifact, catalog *api.CatalogObject) error {
	var metricsViewArtifact MetricsView
	err := mapstructure.Decode(artifact.Definition, &metricsViewArtifact)
	if err != nil {
		return err
	}

	metricsView := &api.MetricsView{}
	err = copier.Copy(metricsView, &metricsViewArtifact)
	if err != nil {
		return err
	}
	// this is needed since measure names are not given by the user
	for i, measure := range metricsView.Measures {
		measure.Name = fmt.Sprintf("measure_%d", i)
	}

	catalog.Name = metricsViewArtifact.Name
	catalog.Type = &api.CatalogObject_MetricsView{
		MetricsView: metricsView,
	}

	if err != nil {
		return err
	}
	return nil
}
