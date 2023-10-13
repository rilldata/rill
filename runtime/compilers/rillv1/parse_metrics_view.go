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

	// Load IANA time zone data
	_ "time/tzdata"
)

// metricsViewYAML is the raw structure of a MetricsView resource defined in YAML
type metricsViewYAML struct {
	commonYAML         `yaml:",inline"` // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	Title              string   `yaml:"title"`
	DisplayName        string   `yaml:"display_name"` // Backwards compatibility
	Description        string   `yaml:"description"`
	Model              string   `yaml:"model"`
	Table              string   `yaml:"table"`
	TimeDimension      string   `yaml:"timeseries"`
	SmallestTimeGrain  string   `yaml:"smallest_time_grain"`
	DefaultTimeRange   string   `yaml:"default_time_range"`
	AvailableTimeZones []string `yaml:"available_time_zones"`
	FirstDayOfWeek     uint32   `yaml:"first_day_of_week"`
	FirstMonthOfYear   uint32   `yaml:"first_month_of_year"`
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
		FormatPreset        string `yaml:"format_preset"`
		FormatD3            string `yaml:"format_d3"`
		Ignore              bool   `yaml:"ignore"`
		ValidPercentOfTotal bool   `yaml:"valid_percent_of_total"`
	}
	Security *struct {
		Access    string `yaml:"access"`
		RowFilter string `yaml:"row_filter"`
		Include   []*struct {
			Names     []string
			Condition string `yaml:"if"`
		}
		Exclude []*struct {
			Names     []string
			Condition string `yaml:"if"`
		}
	}
	DefaultComparison *struct {
		Enabled   bool   `yaml:"enabled"`
		Dimension string `yaml:"dimension"`
		TimeRange string `yaml:"time_range"`
	} `yaml:"default_comparison"`
}

// parseMetricsView parses a metrics view (dashboard) definition and adds the resulting resource to p.Resources.
func (p *Parser) parseMetricsView(ctx context.Context, node *Node) error {
	// Parse YAML
	tmp := &metricsViewYAML{}
	if p.RillYAML != nil && !p.RillYAML.Defaults.MetricsViews.IsZero() {
		if err := p.RillYAML.Defaults.MetricsViews.Decode(tmp); err != nil {
			return pathError{path: node.YAMLPath, err: fmt.Errorf("failed applying defaults from rill.yaml: %w", err)}
		}
	}
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

		if measure.FormatPreset != "" && measure.FormatD3 != "" {
			return fmt.Errorf(`cannot set both "format_preset" and "format_d3" for a measure`)
		}
	}
	if measureCount == 0 {
		return fmt.Errorf("must define at least one measure")
	}

	if tmp.DefaultComparison != nil {
		if tmp.DefaultComparison.Dimension != "" {
			if ok := names[tmp.DefaultComparison.Dimension]; !ok {
				return fmt.Errorf("default time dimension %s doesnt exist", tmp.DefaultComparison.Dimension)
			}
		}

		if tmp.DefaultComparison.TimeRange != "" {
			_, err := duration.ParseISO8601(tmp.DefaultComparison.TimeRange)
			if err != nil {
				return fmt.Errorf("default comparison time range is not valid")
			}
		}
	}

	if tmp.Security != nil {
		templateData := TemplateData{User: map[string]interface{}{
			"name":   "dummy",
			"email":  "mock@example.org",
			"domain": "example.org",
			"groups": []interface{}{"all"},
			"admin":  false,
		}}

		if tmp.Security.Access != "" {
			access, err := ResolveTemplate(tmp.Security.Access, templateData)
			if err != nil {
				return fmt.Errorf(`invalid 'security': 'access' templating is not valid: %w`, err)
			}
			_, err = EvaluateBoolExpression(access)
			if err != nil {
				return fmt.Errorf(`invalid 'security': 'access' expression error: %w`, err)
			}
		}

		if tmp.Security.RowFilter != "" {
			_, err := ResolveTemplate(tmp.Security.RowFilter, templateData)
			if err != nil {
				return fmt.Errorf(`invalid 'security': 'row_filter' templating is not valid: %w`, err)
			}
		}

		if len(tmp.Security.Include) > 0 && len(tmp.Security.Exclude) > 0 {
			return errors.New("invalid 'security': only one of 'include' and 'exclude' can be specified")
		}
		if tmp.Security.Include != nil {
			for _, include := range tmp.Security.Include {
				if include == nil || len(include.Names) == 0 || include.Condition == "" {
					return fmt.Errorf("invalid 'security': 'include' fields must have a valid 'if' condition and 'names' list")
				}
				seen := make(map[string]bool)
				for _, name := range include.Names {
					if seen[name] {
						return fmt.Errorf("invalid 'security': 'include' property %q is duplicated", name)
					}
					seen[name] = true
					if !names[name] {
						return fmt.Errorf("invalid 'security': 'include' property %q does not exists in dimensions or measures list", name)
					}
				}
				cond, err := ResolveTemplate(include.Condition, templateData)
				if err != nil {
					return fmt.Errorf(`invalid 'security': 'if' condition templating for field %q is not valid: %w`, include.Names, err)
				}
				_, err = EvaluateBoolExpression(cond)
				if err != nil {
					return fmt.Errorf(`invalid 'security': 'if' condition for field %q not evaluating to a boolean: %w`, include.Names, err)
				}
			}
		}
		if tmp.Security.Exclude != nil {
			for _, exclude := range tmp.Security.Exclude {
				if exclude == nil || len(exclude.Names) == 0 || exclude.Condition == "" {
					return fmt.Errorf("invalid 'security': 'exclude' fields must have a valid 'if' condition and 'names' list")
				}
				seen := make(map[string]bool)
				for _, name := range exclude.Names {
					if seen[name] {
						return fmt.Errorf("invalid 'security': 'exclude' property %q is duplicated", name)
					}
					seen[name] = true
					if !names[name] {
						return fmt.Errorf("invalid 'security': 'exclude' property %q does not exists in dimensions or measures list", name)
					}
				}
				cond, err := ResolveTemplate(exclude.Condition, templateData)
				if err != nil {
					return fmt.Errorf(`invalid 'security': 'if' condition templating for field %q is not valid: %w`, exclude.Names, err)
				}
				_, err = EvaluateBoolExpression(cond)
				if err != nil {
					return fmt.Errorf(`invalid 'security': 'if' condition for field %q not evaluating to a boolean: %w`, exclude.Names, err)
				}
			}
		}
	}

	node.Refs = append(node.Refs, ResourceName{Name: table})

	r, err := p.insertResource(ResourceKindMetricsView, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.
	spec := r.MetricsViewSpec

	spec.Connector = node.Connector
	spec.Table = table
	spec.Title = tmp.Title
	spec.Description = tmp.Description
	spec.TimeDimension = tmp.TimeDimension
	spec.SmallestTimeGrain = smallestTimeGrain
	spec.DefaultTimeRange = tmp.DefaultTimeRange
	spec.AvailableTimeZones = tmp.AvailableTimeZones
	spec.FirstDayOfWeek = tmp.FirstDayOfWeek
	spec.FirstMonthOfYear = tmp.FirstMonthOfYear

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
			FormatPreset:        measure.FormatPreset,
			FormatD3:            measure.FormatD3,
			ValidPercentOfTotal: measure.ValidPercentOfTotal,
		})
	}

	if tmp.DefaultComparison != nil {
		spec.DefaultComparison = &runtimev1.MetricsViewSpec_DefaultComparison{
			Enabled: tmp.DefaultComparison.Enabled,
		}
		if tmp.DefaultComparison.Dimension != "" {
			spec.DefaultComparison.Dimension = tmp.DefaultComparison.Dimension
		}
		if tmp.DefaultComparison.TimeRange != "" {
			spec.DefaultComparison.TimeRange = tmp.DefaultComparison.TimeRange
		}
	}

	if tmp.Security != nil {
		if spec.Security == nil {
			spec.Security = &runtimev1.MetricsViewSpec_SecurityV2{}
		}
		spec.Security.Access = tmp.Security.Access
		spec.Security.RowFilter = tmp.Security.RowFilter
		// validation has been done above, only one of these will be set
		if tmp.Security.Include != nil {
			for _, include := range tmp.Security.Include {
				spec.Security.Include = append(spec.Security.Include, &runtimev1.MetricsViewSpec_SecurityV2_FieldConditionV2{
					Condition: include.Condition,
					Names:     include.Names,
				})
			}
		}
		if tmp.Security.Exclude != nil {
			for _, exclude := range tmp.Security.Exclude {
				spec.Security.Exclude = append(spec.Security.Exclude, &runtimev1.MetricsViewSpec_SecurityV2_FieldConditionV2{
					Condition: exclude.Condition,
					Names:     exclude.Names,
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
