package yaml

import (
	"github.com/jinzhu/copier"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

/**
 * This file contains the mapping from CatalogObject to Yaml files
 */

const Version = "1.0.0"

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
	Name        string
	Type        string
	Label       string
	Expression  string
	Description string
	Format      string
}

type Dimension struct {
	Name        string
	Type        string
	Label       string
	Description string
	Format      string
}

func toArtifact(catalog *drivers.CatalogObject) (*Artifact, error) {
	artifact := Artifact{
		Version: Version,
		Type:    catalog.Type,
	}

	var err error
	switch catalog.Type {
	case drivers.CatalogObjectTypeSource:
		artifact.Definition, err = toSourceArtifact(catalog)
	case drivers.CatalogObjectTypeModel:
		artifact.Definition, err = toModelArtifact(catalog)
	case drivers.CatalogObjectTypeMetricsView:
		artifact.Definition, err = toMetricsViewArtifact(catalog)
	}
	if err != nil {
		return nil, err
	}

	return &artifact, nil
}

func toSourceArtifact(catalog *drivers.CatalogObject) (*Source, error) {
	var source api.Source
	err := proto.Unmarshal(catalog.Definition, &source)
	if err != nil {
		return nil, err
	}

	return &Source{
		Name:       source.Name,
		Connector:  source.Connector,
		Properties: source.Properties.AsMap(),
	}, nil
}

func toModelArtifact(catalog *drivers.CatalogObject) (*Model, error) {
	var model api.Model
	err := proto.Unmarshal(catalog.Definition, &model)
	if err != nil {
		return nil, err
	}

	modelArtifact := &Model{
		Name: model.Name,
		Sql:  model.Sql,
	}

	switch model.Dialect {
	case api.Model_DuckDB:
		modelArtifact.Dialect = ModelDialectDuckDB
	}

	return modelArtifact, nil
}

func toMetricsViewArtifact(catalog *drivers.CatalogObject) (*MetricsView, error) {
	var metricsView api.MetricsView
	err := proto.Unmarshal(catalog.Definition, &metricsView)
	if err != nil {
		return nil, err
	}

	metricsArtifact := &MetricsView{}
	err = copier.Copy(metricsArtifact, &metricsView)
	if err != nil {
		return nil, err
	}

	return metricsArtifact, nil
}

func fromArtifact(artifact *Artifact) (*drivers.CatalogObject, error) {
	catalog := &drivers.CatalogObject{
		Type: artifact.Type,
	}

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

func fromSourceArtifact(artifact *Artifact, catalog *drivers.CatalogObject) error {
	var sourceArtifact Source
	err := mapstructure.Decode(artifact.Definition, &sourceArtifact)
	if err != nil {
		return err
	}

	catalog.Name = sourceArtifact.Name

	propsPB, err := structpb.NewStruct(sourceArtifact.Properties)
	if err != nil {
		return err
	}
	source := api.Source{
		Name:       catalog.Name,
		Connector:  sourceArtifact.Connector,
		Properties: propsPB,
	}

	catalog.Definition, err = proto.Marshal(&source)
	if err != nil {
		return err
	}

	return nil
}

func fromModelArtifact(artifact *Artifact, catalog *drivers.CatalogObject) error {
	var modelArtifact Model
	err := mapstructure.Decode(artifact.Definition, &modelArtifact)
	if err != nil {
		return err
	}

	catalog.Name = modelArtifact.Name

	model := api.Model{
		Name: modelArtifact.Name,
		Sql:  modelArtifact.Sql,
	}
	switch modelArtifact.Dialect {
	case ModelDialectDuckDB:
		model.Dialect = api.Model_DuckDB
	}

	catalog.Definition, err = proto.Marshal(&model)
	if err != nil {
		return err
	}

	return nil
}

func fromMetricsViewArtifact(artifact *Artifact, catalog *drivers.CatalogObject) error {
	var metricsViewArtifact MetricsView
	err := mapstructure.Decode(artifact.Definition, &metricsViewArtifact)
	if err != nil {
		return err
	}

	catalog.Name = metricsViewArtifact.Name

	metricsView := api.MetricsView{}
	err = copier.Copy(&metricsView, &metricsViewArtifact)
	if err != nil {
		return err
	}

	catalog.Definition, err = proto.Marshal(&metricsView)
	if err != nil {
		return err
	}
	return nil
}
