package yaml

import (
	"fmt"

	"github.com/jinzhu/copier"
	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"google.golang.org/protobuf/types/known/structpb"
)

/**
 * This file contains the mapping from CatalogObject to Yaml files
 */
type Source struct {
	Type                  string
	Path                  string         `yaml:"path,omitempty"`
	CsvDelimiter          string         `yaml:"csv.delimiter,omitempty" mapstructure:"csv.delimiter,omitempty"`
	URI                   string         `yaml:"uri,omitempty"`
	Region                string         `yaml:"region,omitempty" mapstructure:"region,omitempty"`
	GlobMaxTotalSize      int64          `yaml:"glob.max_total_size,omitempty" mapstructure:"glob.max_total_size,omitempty"`
	GlobMaxObjectsMatched int            `yaml:"glob.max_objects_matched,omitempty" mapstructure:"glob.max_objects_matched,omitempty"`
	GlobMaxObjectsListed  int64          `yaml:"glob.max_objects_listed,omitempty" mapstructure:"glob.max_objects_listed,omitempty"`
	GlobPageSize          int            `yaml:"glob.page_size,omitempty" mapstructure:"glob.page_size,omitempty"`
	HivePartition         *bool          `yaml:"hive_partitioning,omitempty" mapstructure:"hive_partitioning,omitempty"`
	Policy                *ExtractPolicy `yaml:"extract,omitempty" mapstructure:"source.extract,omitempty"`
}

type ExtractPolicy struct {
	Row       *ExtractConfig `yaml:"rows,omitempty" mapstructure:"rows,omitempty" json:"rows,omitempty"`
	Partition *ExtractConfig `yaml:"partitions,omitempty" mapstructure:"partitions,omitempty" json:"partitions,omitempty"`
}

type ExtractConfig struct {
	Strategy string `yaml:"strategy,omitempty" mapstructure:"strategy,omitempty" json:"strategy,omitempty"`
	Size     string `yaml:"size,omitempty" mapstructure:"size,omitempty" json:"size,omitempty"`
}

type MetricsView struct {
	Label            string `yaml:"display_name"`
	Description      string
	Model            string
	TimeDimension    string `yaml:"timeseries"`
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
	Ignore      bool   `yaml:"ignore,omitempty"`
}

type Dimension struct {
	Label       string
	Property    string `copier:"Name"`
	Description string
	Ignore      bool `yaml:"ignore,omitempty"`
}

func toSourceArtifact(catalog *drivers.CatalogEntry) (*Source, error) {
	source := &Source{
		Type: catalog.GetSource().Connector,
	}

	props := catalog.GetSource().Properties.AsMap()

	err := mapstructure.Decode(props, source)
	if err != nil {
		return nil, err
	}

	if source.Path != "" && catalog.GetSource().Connector != "local_file" {
		source.URI = source.Path
		source.Path = ""
	}

	extract, err := toExtractArtifact(catalog.GetSource().GetPolicy())
	if err != nil {
		return nil, err
	}

	source.Policy = extract
	return source, nil
}

func toExtractArtifact(extract *runtimev1.Source_ExtractPolicy) (*ExtractPolicy, error) {
	if extract == nil {
		return nil, nil
	}

	sourceExtract := &ExtractPolicy{}
	err := copier.Copy(sourceExtract, extract)
	if err != nil {
		return nil, err
	}

	return sourceExtract, nil
}

func toMetricsViewArtifact(catalog *drivers.CatalogEntry) (*MetricsView, error) {
	metricsArtifact := &MetricsView{}
	err := copier.Copy(metricsArtifact, catalog.Object)
	if err != nil {
		return nil, err
	}

	return metricsArtifact, nil
}

func fromSourceArtifact(source *Source, path string) (*drivers.CatalogEntry, error) {
	props := map[string]interface{}{}
	if source.Type == "local_file" {
		props["path"] = source.Path
	} else {
		props["path"] = source.URI
	}
	if source.Region != "" {
		props["region"] = source.Region
	}
	if source.CsvDelimiter != "" {
		props["csv.delimiter"] = source.CsvDelimiter
	}
	if source.GlobMaxTotalSize != 0 {
		props["glob.max_total_size"] = source.GlobMaxTotalSize
	}

	if source.GlobMaxObjectsMatched != 0 {
		props["glob.max_objects_matched"] = source.GlobMaxObjectsMatched
	}

	if source.GlobMaxObjectsListed != 0 {
		props["glob.max_objects_listed"] = source.GlobMaxObjectsListed
	}

	if source.GlobPageSize != 0 {
		props["glob.page_size"] = source.GlobPageSize
	}

	if source.HivePartition != nil {
		props["hive_partitioning"] = *source.HivePartition
	}

	propsPB, err := structpb.NewStruct(props)
	if err != nil {
		return nil, err
	}

	extract, err := fromExtractArtifact(source.Policy)
	if err != nil {
		return nil, err
	}

	name := fileutil.Stem(path)
	return &drivers.CatalogEntry{
		Name: name,
		Type: drivers.ObjectTypeSource,
		Path: path,
		Object: &runtimev1.Source{
			Name:       name,
			Connector:  source.Type,
			Properties: propsPB,
			Policy:     extract,
		},
	}, nil
}

func fromExtractArtifact(sourceExtract *ExtractPolicy) (*runtimev1.Source_ExtractPolicy, error) {
	if sourceExtract == nil {
		return nil, nil
	}

	extractPolicy := &runtimev1.Source_ExtractPolicy{}
	err := copier.Copy(extractPolicy, sourceExtract)
	if err != nil {
		return nil, err
	}

	return extractPolicy, nil
}

func fromMetricsViewArtifact(metrics *MetricsView, path string) (*drivers.CatalogEntry, error) {
	// remove ignored measures and dimensions
	var measures []*Measure
	for _, measure := range metrics.Measures {
		if measure.Ignore {
			continue
		}
		measures = append(measures, measure)
	}
	metrics.Measures = measures

	var dimensions []*Dimension
	for _, dimension := range metrics.Dimensions {
		if dimension.Ignore {
			continue
		}
		dimensions = append(dimensions, dimension)
	}
	metrics.Dimensions = dimensions

	apiMetrics := &runtimev1.MetricsView{}
	err := copier.Copy(apiMetrics, metrics)
	if err != nil {
		return nil, err
	}

	// this is needed since measure names are not given by the user
	for i, measure := range apiMetrics.Measures {
		measure.Name = fmt.Sprintf("measure_%d", i)
	}

	name := fileutil.Stem(path)
	apiMetrics.Name = name
	return &drivers.CatalogEntry{
		Name:   name,
		Type:   drivers.ObjectTypeMetricsView,
		Path:   path,
		Object: apiMetrics,
	}, nil
}
