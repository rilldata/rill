package rillv1

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/duration"
	"gopkg.in/yaml.v3"
)

// metricsViewYAML is the raw structure of a MetricsView resource defined in YAML
type metricsViewYAML struct {
	commonYAML         `yaml:",inline"` // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	Title              string           `yaml:"title"`
	DisplayName        string           `yaml:"display_name"` // Backwards compatibility
	Description        string           `yaml:"description"`
	Model              string           `yaml:"model"`
	Table              string           `yaml:"table"`
	TimeDimension      string           `yaml:"timeseries"`
	SmallestTimeGrain  string           `yaml:"smallest_time_grain"`
	DefaultTimeRange   string           `yaml:"default_time_range"`
	AvailableTimeZones []string         `yaml:"available_time_zones"`
	Dimensions         []*struct {
		Name        string
		Label       string
		Column      string
		Property    string // For backwards compatibility
		Description string
		Ignore      bool `yaml:"ignore"`
	}
	Measures []*struct {
		Name                string
		Label               string
		Expression          string
		Description         string
		Format              string `yaml:"format_preset"`
		Ignore              bool   `yaml:"ignore"`
		ValidPercentOfTotal bool   `yaml:"valid_percent_of_total"`
	}
	Policy *struct {
		HasAccess string `yaml:"has_access"`
		Filter    string `yaml:"filter"`
		Include   []*struct {
			Name      string
			Condition string `yaml:"if"`
		}
		Exclude []*struct {
			Name      string
			Condition string `yaml:"if"`
		}
	}
}

// parseMetricsView parses a metrics view (dashboard) definition and adds the resulting resource to p.Resources.
func (p *Parser) parseMetricsView(ctx context.Context, node *Node) error {
	// Parse YAML
	tmp := &metricsViewYAML{}
	if node.YAMLRaw != "" {
		// Can't use node.YAML because we need to set KnownFields for metrics views
		dec := yaml.NewDecoder(strings.NewReader(node.YAMLRaw))
		dec.KnownFields(true)
		if err := dec.Decode(tmp); err != nil {
			return pathError{path: node.YAMLPath, err: newYAMLError(err)}
		}
	}

	// Backwards compatibility
	if tmp.DisplayName != "" && tmp.Title == "" {
		tmp.Title = tmp.DisplayName
	}

	var table string
	if tmp.Table == "" {
		table = tmp.Model
	} else if tmp.Model == "" {
		table = tmp.Table
	} else {
		return fmt.Errorf(`cannot set both the "model" field and the "table" field`)
	}
	if table == "" {
		return fmt.Errorf(`must set a value for either the "model" field or the "table" field`)
	}

	smallestTimeGrain, err := parseTimeGrain(tmp.SmallestTimeGrain)
	if err != nil {
		return fmt.Errorf(`invalid "smallest_time_grain": %w`, err)
	}

	if tmp.DefaultTimeRange != "" {
		_, err := duration.ParseISO8601(tmp.DefaultTimeRange)
		if err != nil {
			return fmt.Errorf(`invalid "default_time_range": %w`, err)
		}
	}

	for _, tz := range tmp.AvailableTimeZones {
		_, err := time.LoadLocation(tz)
		if err != nil {
			return err
		}
	}

	names := make(map[string]bool)
	columns := make(map[string]bool)
	for i, dim := range tmp.Dimensions {
		if dim.Ignore {
			continue
		}

		// Backwards compatibility
		if dim.Property != "" && dim.Column == "" {
			dim.Column = dim.Property
		}

		// Backwards compatibility
		if dim.Name == "" {
			if dim.Column == "" {
				dim.Name = fmt.Sprintf("dimension_%d", i)
			} else {
				dim.Name = dim.Column
			}
		}

		lower := strings.ToLower(dim.Name)
		if ok := names[lower]; ok {
			return fmt.Errorf("found duplicate dimension or measure name %q", dim.Name)
		}
		names[lower] = true

		lower = strings.ToLower(dim.Column)
		if ok := columns[lower]; ok {
			return fmt.Errorf("found duplicate dimension column name %q", dim.Column)
		}
		columns[lower] = true
	}

	measureCount := 0
	for i, measure := range tmp.Measures {
		if measure.Ignore {
			continue
		}

		measureCount++

		// Backwards compatibility
		if measure.Name == "" {
			measure.Name = fmt.Sprintf("measure_%d", i)
		}

		lower := strings.ToLower(measure.Name)
		if ok := names[lower]; ok {
			return fmt.Errorf("found duplicate dimension or measure name %q", measure.Name)
		}
		names[lower] = true

		if ok := columns[lower]; ok {
			return fmt.Errorf("measure name %q coincides with a dimension column name", measure.Name)
		}
	}
	if measureCount == 0 {
		return fmt.Errorf("must define at least one measure")
	}

	if tmp.Policy != nil {
		if len(tmp.Policy.Include) > 0 && len(tmp.Policy.Exclude) > 0 {
			return errors.New("invalid 'policy': only one of 'include' and 'exclude' can be specified")
		}
		if tmp.Policy.Include != nil {
			for _, include := range tmp.Policy.Include {
				if include.Name == "" || include.Condition == "" {
					return fmt.Errorf("invalid 'policy': 'include' fields must have a valid 'name' and 'if' condition")
				}
				if !names[include.Name] {
					return fmt.Errorf("invalid 'policy': 'include' property %q does not exists in dimensions list", include.Name)
				}
			}
		}
		if tmp.Policy.Exclude != nil {
			for _, exclude := range tmp.Policy.Exclude {
				if exclude.Name == "" || exclude.Condition == "" {
					return fmt.Errorf("invalid 'policy': 'include' fields must have a valid 'name' and 'if' condition")
				}
				if !names[exclude.Name] {
					return fmt.Errorf("invalid 'policy': 'exclude' property %q does not exists in dimensions list", exclude.Name)
				}
			}
		}
	}

	node.Refs = append(node.Refs, ResourceName{Name: table})

	// NOTE: After calling upsertResource, an error must not be returned. Any validation should be done before calling it.
	r := p.upsertResource(ResourceKindMetricsView, node.Name, node.Paths, node.Refs...)
	spec := r.MetricsViewSpec

	spec.Connector = node.Connector
	spec.Table = table
	spec.Title = tmp.Title
	spec.Description = tmp.Description
	spec.TimeDimension = tmp.TimeDimension
	spec.SmallestTimeGrain = smallestTimeGrain
	spec.DefaultTimeRange = tmp.DefaultTimeRange
	spec.AvailableTimeZones = tmp.AvailableTimeZones

	for _, dim := range tmp.Dimensions {
		if dim.Ignore {
			continue
		}

		spec.Dimensions = append(spec.Dimensions, &runtimev1.MetricsViewSpec_DimensionV2{
			Name:        dim.Name,
			Column:      dim.Column,
			Label:       dim.Label,
			Description: dim.Description,
		})
	}

	for _, measure := range tmp.Measures {
		if measure.Ignore {
			continue
		}

		spec.Measures = append(spec.Measures, &runtimev1.MetricsViewSpec_MeasureV2{
			Name:                measure.Name,
			Expression:          measure.Expression,
			Label:               measure.Label,
			Description:         measure.Description,
			Format:              measure.Format,
			ValidPercentOfTotal: measure.ValidPercentOfTotal,
		})
	}

	if tmp.Policy != nil {
		if spec.Policy == nil {
			spec.Policy = &runtimev1.MetricsViewSpec_PolicyV2{}
		}
		spec.Policy.HasAccess = tmp.Policy.HasAccess
		spec.Policy.Filter = tmp.Policy.Filter
		// validation has been done above, only one of these will be set
		if tmp.Policy.Include != nil {
			for _, include := range tmp.Policy.Include {
				spec.Policy.Include = append(spec.Policy.Include, &runtimev1.MetricsViewSpec_PolicyV2_FieldConditionV2{
					Name:      include.Name,
					Condition: include.Condition,
				})
			}
		}
		if tmp.Policy.Exclude != nil {
			for _, exclude := range tmp.Policy.Exclude {
				spec.Policy.Exclude = append(spec.Policy.Exclude, &runtimev1.MetricsViewSpec_PolicyV2_FieldConditionV2{
					Name:      exclude.Name,
					Condition: exclude.Condition,
				})
			}
		}
	}

	return nil
}

// parseTimeGrain parses a YAML time grain string
func parseTimeGrain(s string) (runtimev1.TimeGrain, error) {
	switch strings.ToLower(s) {
	case "":
		return runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED, nil
	case "ms", "millisecond":
		return runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND, nil
	case "s", "second":
		return runtimev1.TimeGrain_TIME_GRAIN_SECOND, nil
	case "min", "minute":
		return runtimev1.TimeGrain_TIME_GRAIN_MINUTE, nil
	case "h", "hour":
		return runtimev1.TimeGrain_TIME_GRAIN_HOUR, nil
	case "d", "day":
		return runtimev1.TimeGrain_TIME_GRAIN_DAY, nil
	case "w", "week":
		return runtimev1.TimeGrain_TIME_GRAIN_WEEK, nil
	case "month":
		return runtimev1.TimeGrain_TIME_GRAIN_MONTH, nil
	case "q", "quarter":
		return runtimev1.TimeGrain_TIME_GRAIN_QUARTER, nil
	case "y", "year":
		return runtimev1.TimeGrain_TIME_GRAIN_YEAR, nil
	default:
		return runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED, fmt.Errorf("invalid time grain %q", s)
	}
}
