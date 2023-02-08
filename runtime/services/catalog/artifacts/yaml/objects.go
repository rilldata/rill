package yaml

import (
	"fmt"
	"strings"

	"github.com/senseyeio/duration"

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
	Path                  string `yaml:"path,omitempty"`
	CsvDelimiter          string `yaml:"csv.delimiter,omitempty" mapstructure:"csv.delimiter,omitempty"`
	URI                   string `yaml:"uri,omitempty"`
	Region                string `yaml:"region,omitempty" mapstructure:"region,omitempty"`
	GlobMaxTotalSize      int64  `yaml:"glob.max_total_size,omitempty" mapstructure:"glob.max_total_size,omitempty"`
	GlobMaxObjectsMatched int    `yaml:"glob.max_objects_matched,omitempty" mapstructure:"glob.max_objects_matched,omitempty"`
	GlobMaxObjectsListed  int64  `yaml:"glob.max_objects_listed,omitempty" mapstructure:"glob.max_objects_listed,omitempty"`
	GlobPageSize          int    `yaml:"glob.page_size,omitempty" mapstructure:"glob.page_size,omitempty"`
}

type MetricsView struct {
	Label            string `yaml:"display_name"`
	Description      string
	Model            string
	TimeDimension    string   `yaml:"timeseries"`
	TimeGrains       []string `yaml:"time_grains"`
	DefaultTimeGrain string   `yaml:"default_time_grain"`
	DefaultTimeRange string   `yaml:"default_time_range"`
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

	return source, nil
}

func toMetricsViewArtifact(catalog *drivers.CatalogEntry) (*MetricsView, error) {
	metricsArtifact := &MetricsView{}
	err := copier.Copy(metricsArtifact, catalog.Object)
	var timeGrains []string
	for _, timeGrainEnum := range catalog.GetMetricsView().TimeGrains {
		timeGrains = append(timeGrains, getTimeGrainString(timeGrainEnum))
	}
	metricsArtifact.TimeGrains = timeGrains
	metricsArtifact.DefaultTimeGrain = getTimeGrainString(catalog.GetMetricsView().DefaultTimeGrain)
	metricsArtifact.DefaultTimeRange = catalog.GetMetricsView().DefaultTimeRange
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
	propsPB, err := structpb.NewStruct(props)
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
		},
	}, nil
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

	// validate correctness of default time range
	if metrics.DefaultTimeRange != "" {
		_, err := duration.ParseISO8601(metrics.DefaultTimeRange)
		if err != nil {
			return nil, fmt.Errorf("invalid default_time_grain: %w", err)
		}
		apiMetrics.DefaultTimeRange = metrics.DefaultTimeRange
	}

	err := copier.Copy(apiMetrics, metrics)
	if err != nil {
		return nil, err
	}

	// this is needed since measure names are not given by the user
	for i, measure := range apiMetrics.Measures {
		measure.Name = fmt.Sprintf("measure_%d", i)
	}

	for _, timeGrain := range metrics.TimeGrains {
		apiMetrics.TimeGrains = append(apiMetrics.TimeGrains, getTimeGrainEnum(timeGrain))
	}
	apiMetrics.DefaultTimeGrain = getTimeGrainEnum(metrics.DefaultTimeGrain)

	name := fileutil.Stem(path)
	apiMetrics.Name = name
	return &drivers.CatalogEntry{
		Name:   name,
		Type:   drivers.ObjectTypeMetricsView,
		Path:   path,
		Object: apiMetrics,
	}, nil
}

// Get TimeGrain enum from string
func getTimeGrainEnum(timeGrain string) runtimev1.TimeGrain {
	switch strings.ToLower(timeGrain) {
	case "millisecond":
		return runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND
	case "second":
		return runtimev1.TimeGrain_TIME_GRAIN_SECOND
	case "minute":
		return runtimev1.TimeGrain_TIME_GRAIN_MINUTE
	case "hour":
		return runtimev1.TimeGrain_TIME_GRAIN_HOUR
	case "day":
		return runtimev1.TimeGrain_TIME_GRAIN_DAY
	case "week":
		return runtimev1.TimeGrain_TIME_GRAIN_WEEK
	case "month":
		return runtimev1.TimeGrain_TIME_GRAIN_MONTH
	case "year":
		return runtimev1.TimeGrain_TIME_GRAIN_YEAR
	default:
		return runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED
	}
}

// Get TimeGrain string from enum
func getTimeGrainString(timeGrain runtimev1.TimeGrain) string {
	switch timeGrain {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		return "millisecond"
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return "second"
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return "minute"
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		return "hour"
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		return "day"
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		return "week"
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		return "month"
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		return "year"
	default:
		return ""
	}
}
