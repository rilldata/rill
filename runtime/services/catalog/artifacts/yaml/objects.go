package yaml

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/duration"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"google.golang.org/protobuf/types/known/structpb"

	// Load IANA time zone data
	_ "time/tzdata"
)

/**
 * This file contains the mapping from CatalogObject to Yaml files
 */
type Source struct {
	Type                        string
	Path                        string         `yaml:"path,omitempty"`
	CsvDelimiter                string         `yaml:"csv.delimiter,omitempty" mapstructure:"csv.delimiter,omitempty"`
	URI                         string         `yaml:"uri,omitempty"`
	Region                      string         `yaml:"region,omitempty" mapstructure:"region,omitempty"`
	S3Endpoint                  string         `yaml:"endpoint,omitempty" mapstructure:"endpoint,omitempty"`
	GlobMaxTotalSize            int64          `yaml:"glob.max_total_size,omitempty" mapstructure:"glob.max_total_size,omitempty"`
	GlobMaxObjectsMatched       int            `yaml:"glob.max_objects_matched,omitempty" mapstructure:"glob.max_objects_matched,omitempty"`
	GlobMaxObjectsListed        int64          `yaml:"glob.max_objects_listed,omitempty" mapstructure:"glob.max_objects_listed,omitempty"`
	GlobPageSize                int            `yaml:"glob.page_size,omitempty" mapstructure:"glob.page_size,omitempty"`
	BatchSize                   string         `yaml:"batch_size,omitempty" mapstructure:"batch_size,omitempty"`
	HivePartition               *bool          `yaml:"hive_partitioning,omitempty" mapstructure:"hive_partitioning,omitempty"`
	Timeout                     int32          `yaml:"timeout,omitempty"`
	Format                      string         `yaml:"format,omitempty" mapstructure:"format,omitempty"`
	Extract                     map[string]any `yaml:"extract,omitempty" mapstructure:"extract,omitempty"`
	DuckDBProps                 map[string]any `yaml:"duckdb,omitempty" mapstructure:"duckdb,omitempty"`
	Headers                     map[string]any `yaml:"headers,omitempty" mapstructure:"headers,omitempty"`
	AllowSchemaRelaxation       *bool          `yaml:"allow_schema_relaxation,omitempty" mapstructure:"allow_schema_relaxation,omitempty"`
	IngestAllowSchemaRelaxation *bool          `yaml:"ingest.allow_schema_relaxation,omitempty" mapstructure:"ingest.allow_schema_relaxation,omitempty"`
	SQL                         string         `yaml:"sql,omitempty" mapstructure:"sql,omitempty"`
	DB                          string         `yaml:"db,omitempty" mapstructure:"db,omitempty"`
	ProjectID                   string         `yaml:"project_id,omitempty" mapstructure:"project_id,omitempty"`
}

type MetricsView struct {
	Label              string `yaml:"title"`
	DisplayName        string `yaml:"display_name,omitempty"` // for backwards compatibility
	Description        string
	Model              string
	TimeDimension      string   `yaml:"timeseries"`
	SmallestTimeGrain  string   `yaml:"smallest_time_grain"`
	DefaultTimeRange   string   `yaml:"default_time_range"`
	AvailableTimeZones []string `yaml:"available_time_zones,omitempty"`
	Dimensions         []*Dimension
	Measures           []*Measure
	Security           *Security `yaml:"security,omitempty"`
}

type Security struct {
	Access    string              `yaml:"access,omitempty"`
	RowFilter string              `yaml:"row_filter,omitempty"`
	Include   []*ConditionalField `yaml:"include,omitempty"`
	Exclude   []*ConditionalField `yaml:"exclude,omitempty"`
}

type ConditionalField struct {
	Names     []string
	Condition string `yaml:"if"`
}

type Measure struct {
	Label               string
	Name                string
	Expression          string
	Description         string
	Format              string `yaml:"format_preset"`
	Ignore              bool   `yaml:"ignore,omitempty"`
	ValidPercentOfTotal bool   `yaml:"valid_percent_of_total,omitempty"`
}

type Dimension struct {
	Name        string
	Label       string
	Property    string `yaml:"property,omitempty"`
	Column      string
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
	metricsArtifact.SmallestTimeGrain = getTimeGrainString(catalog.GetMetricsView().SmallestTimeGrain)
	metricsArtifact.DefaultTimeRange = catalog.GetMetricsView().DefaultTimeRange
	if err != nil {
		return nil, err
	}

	return metricsArtifact, nil
}

func fromSourceArtifact(source *Source, path string) (*drivers.CatalogEntry, error) {
	props := map[string]interface{}{}

	if source.Path != "" {
		props["path"] = source.Path
	}

	if source.URI != "" {
		props["uri"] = source.URI
	}

	if source.Region != "" {
		props["region"] = source.Region
	}

	if source.Extract != nil {
		props["extract"] = source.Extract
	}

	if source.DuckDBProps != nil {
		props["duckdb"] = source.DuckDBProps
	}

	if source.CsvDelimiter != "" {
		props["csv.delimiter"] = source.CsvDelimiter
	}

	if source.HivePartition != nil {
		props["hive_partitioning"] = *source.HivePartition
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

	if source.BatchSize != "" {
		props["batch_size"] = source.BatchSize
	}

	if source.S3Endpoint != "" {
		props["endpoint"] = source.S3Endpoint
	}

	if source.Format != "" {
		props["format"] = source.Format
	}

	if source.Headers != nil {
		props["headers"] = source.Headers
	}

	if source.AllowSchemaRelaxation != nil {
		props["allow_schema_relaxation"] = *source.AllowSchemaRelaxation
	}

	if source.IngestAllowSchemaRelaxation != nil {
		props["ingest.allow_schema_relaxation"] = *source.IngestAllowSchemaRelaxation
	}

	if source.SQL != "" {
		props["sql"] = source.SQL
	}

	if source.DB != "" {
		props["db"] = source.DB
	}

	if source.ProjectID != "" {
		props["project_id"] = source.ProjectID
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
			Name:           name,
			Connector:      source.Type,
			Properties:     propsPB,
			TimeoutSeconds: source.Timeout,
		},
	}, nil
}

func fromMetricsViewArtifact(metrics *MetricsView, path string) (*drivers.CatalogEntry, error) {
	if metrics.DisplayName != "" && metrics.Label == "" {
		// backwards compatibility
		metrics.Label = metrics.DisplayName
	}

	names := map[string]bool{}

	// remove ignored measures and dimensions
	var measures []*Measure
	for _, measure := range metrics.Measures {
		if measure.Ignore {
			continue
		}
		measures = append(measures, measure)
		if measure.Name != "" {
			names[measure.Name] = true
		}
	}
	metrics.Measures = measures

	var dimensions []*Dimension
	for _, dimension := range metrics.Dimensions {
		if dimension.Ignore {
			continue
		}
		if dimension.Property != "" && dimension.Column == "" {
			// backwards compatibility when we were using `property` instead of `column`
			dimension.Column = dimension.Property
		}
		dimensions = append(dimensions, dimension)

		if dimension.Name != "" {
			names[dimension.Name] = true
		} else if dimension.Column != "" {
			names[dimension.Column] = true
		}
	}
	metrics.Dimensions = dimensions

	if metrics.Security != nil {
		templateData := rillv1.TemplateData{User: map[string]interface{}{
			"name":   "dummy",
			"email":  "mock@example.org",
			"domain": "example.org",
			"groups": []interface{}{"all"},
			"admin":  false,
		}}

		if metrics.Security.Access != "" {
			access, err := rillv1.ResolveTemplate(metrics.Security.Access, templateData)
			if err != nil {
				return nil, fmt.Errorf(`invalid 'security': 'access' templating is not valid: %w`, err)
			}
			_, err = rillv1.EvaluateBoolExpression(access)
			if err != nil {
				return nil, fmt.Errorf(`invalid 'security': 'access' expression error: %w`, err)
			}
		}

		if metrics.Security.RowFilter != "" {
			_, err := rillv1.ResolveTemplate(metrics.Security.RowFilter, templateData)
			if err != nil {
				return nil, fmt.Errorf(`invalid 'security': 'row_filter' templating is not valid: %w`, err)
			}
		}

		if len(metrics.Security.Include) > 0 && len(metrics.Security.Exclude) > 0 {
			return nil, errors.New("invalid 'security': only one of 'include' and 'exclude' can be specified")
		}

		err := validatedPolicyFieldList(metrics.Security.Include, names, "include", templateData)
		if err != nil {
			return nil, err
		}

		err = validatedPolicyFieldList(metrics.Security.Exclude, names, "exclude", templateData)
		if err != nil {
			return nil, err
		}
	}

	apiMetrics := &runtimev1.MetricsView{}

	// validate correctness of default time range
	if metrics.DefaultTimeRange != "" {
		_, err := duration.ParseISO8601(metrics.DefaultTimeRange)
		if err != nil {
			return nil, fmt.Errorf("invalid default_time_range: %w", err)
		}
		apiMetrics.DefaultTimeRange = metrics.DefaultTimeRange
	}

	// validate time zone locations
	for _, tz := range metrics.AvailableTimeZones {
		_, err := time.LoadLocation(tz)
		if err != nil {
			return nil, err
		}
	}

	err := copier.Copy(apiMetrics, metrics)
	if err != nil {
		return nil, err
	}

	// this is needed since measure names are not given by the user
	for i, measure := range apiMetrics.Measures {
		if measure.Name == "" {
			measure.Name = fmt.Sprintf("measure_%d", i)
		}
	}

	// backwards compatibility where name was used as property
	for i, dimension := range apiMetrics.Dimensions {
		if dimension.Name == "" {
			if dimension.Column == "" {
				// if there is no name and property add dimension_<index> as name
				dimension.Name = fmt.Sprintf("dimension_%d", i)
			} else {
				// else use property as name
				dimension.Name = dimension.Column
			}
		}
	}

	timeGrainEnum, err := getTimeGrainEnum(metrics.SmallestTimeGrain)
	if err != nil {
		return nil, err
	}
	apiMetrics.SmallestTimeGrain = timeGrainEnum

	name := fileutil.Stem(path)
	apiMetrics.Name = name
	return &drivers.CatalogEntry{
		Name:   name,
		Type:   drivers.ObjectTypeMetricsView,
		Path:   path,
		Object: apiMetrics,
	}, nil
}

func validatedPolicyFieldList(fieldConditions []*ConditionalField, names map[string]bool, property string, templateData rillv1.TemplateData) error {
	if len(fieldConditions) > 0 {
		for _, field := range fieldConditions {
			if field == nil || len(field.Names) == 0 || field.Condition == "" {
				return fmt.Errorf("invalid 'security': '%s' fields must have a valid 'if' condition and 'names' list", property)
			}
			seen := map[string]bool{}
			for _, name := range field.Names {
				if seen[name] {
					return fmt.Errorf("invalid 'security': '%s' property %q is duplicated", property, name)
				}
				seen[name] = true
				if !names[name] {
					return fmt.Errorf("invalid 'security': '%s' property %q does not exists in dimensions or measures list", property, name)
				}
			}
			cond, err := rillv1.ResolveTemplate(field.Condition, templateData)
			if err != nil {
				return fmt.Errorf(`invalid 'security': 'if' condition templating for field %q is not valid: %w`, field.Names, err)
			}
			_, err = rillv1.EvaluateBoolExpression(cond)
			if err != nil {
				return fmt.Errorf(`invalid 'security': 'if' condition for field %q not valuating to a boolean: %w`, field.Names, err)
			}
		}
	}
	return nil
}

// Get TimeGrain enum from string
func getTimeGrainEnum(timeGrain string) (runtimev1.TimeGrain, error) {
	switch strings.ToLower(timeGrain) {
	case "":
		return runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED, nil
	case "millisecond":
		return runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND, nil
	case "second":
		return runtimev1.TimeGrain_TIME_GRAIN_SECOND, nil
	case "minute":
		return runtimev1.TimeGrain_TIME_GRAIN_MINUTE, nil
	case "hour":
		return runtimev1.TimeGrain_TIME_GRAIN_HOUR, nil
	case "day":
		return runtimev1.TimeGrain_TIME_GRAIN_DAY, nil
	case "week":
		return runtimev1.TimeGrain_TIME_GRAIN_WEEK, nil
	case "month":
		return runtimev1.TimeGrain_TIME_GRAIN_MONTH, nil
	case "quarter":
		return runtimev1.TimeGrain_TIME_GRAIN_QUARTER, nil
	case "year":
		return runtimev1.TimeGrain_TIME_GRAIN_YEAR, nil
	default:
		return runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED, fmt.Errorf("invalid time grain: %s", timeGrain)
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
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		return "quarter"
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		return "year"
	default:
		return ""
	}
}
