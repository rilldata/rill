package rillv1

import (
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"gopkg.in/yaml.v3"
)

// FieldSelectorYAML parses a list of names with support for a '*' scalar for all names,
// and support for a nested "exclude:" list for selecting all except the listed names.
//
// Note that '*' is represented by setting Exclude to true and leaving Names nil.
// (Because excluding nothing is the same as including everything.)
type FieldSelectorYAML struct {
	All              bool
	Fields           *[]string
	Regex            string
	DuckDBExpression string
	Invert           bool
}

func (y *FieldSelectorYAML) UnmarshalYAML(v *yaml.Node) error {
	if v == nil {
		return nil
	}
	switch v.Kind {
	case yaml.ScalarNode:
		if v.Value == "*" {
			y.All = true
			return nil
		}
		return fmt.Errorf("unexpected scalar %q", v.Value)
	case yaml.SequenceNode:
		fields := make([]string, len(v.Content))
		for i, c := range v.Content {
			if c.Kind != yaml.ScalarNode {
				return fmt.Errorf("unexpected non-string list entry on line %d", c.Line)
			}
			fields[i] = c.Value
		}
		y.Fields = &fields
	case yaml.MappingNode:
		tmp := &struct {
			Regex   string    `yaml:"regex"`
			Expr    string    `yaml:"expr"`
			Exclude yaml.Node `yaml:"exclude"`
		}{}
		err := v.Decode(tmp)
		if err != nil {
			return err
		}

		n := 0
		if tmp.Regex != "" {
			n++
			y.Regex = tmp.Regex
		}
		if tmp.Expr != "" {
			n++
			y.DuckDBExpression = tmp.Expr
		}
		if !tmp.Exclude.IsZero() {
			n++
			// Exclude has the same options, just nested one level.
			// For simpliciy, we can just recurse on it and invert the result.
			err = y.UnmarshalYAML(&tmp.Exclude)
			if err != nil {
				return fmt.Errorf("error parsing `exclude` field: %w", err)
			}
			y.Invert = !y.Invert // Oh the irony
		}
		if n != 1 {
			return errors.New("expected one of '*', list of names, `regex` field, `expr` field or `exclude` field")
		}
	default:
		return fmt.Errorf("expected one of '*', list of names, `regex` field, `expr` field or `exclude` field, got type %q", v.Kind)
	}
	return nil
}

// TryResolve attempts to resolve the field selector to a list of fields without any further context.
// It returns false if the field selector requires context about which fields are available.
func (y *FieldSelectorYAML) TryResolve() ([]string, bool) {
	if y != nil && !y.Invert && y.Fields != nil {
		return *y.Fields, true
	}
	return nil, false
}

// Proto returns the protocol buffer representation of a FieldSelector.
// It is recommended only to use this if TryResolve cannot return a list of fields outright.
func (y *FieldSelectorYAML) Proto() *runtimev1.FieldSelector {
	// If not specified, default to '*' (include all).
	if y == nil {
		return &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}}
	}

	// Build output
	res := &runtimev1.FieldSelector{Invert: y.Invert}
	if y.All {
		res.Selector = &runtimev1.FieldSelector_All{All: true}
	} else if y.Fields != nil {
		res.Selector = &runtimev1.FieldSelector_Fields{Fields: &runtimev1.StringListValue{Values: *y.Fields}}
	} else if y.Regex != "" {
		res.Selector = &runtimev1.FieldSelector_Regex{Regex: y.Regex}
	} else if y.DuckDBExpression != "" {
		res.Selector = &runtimev1.FieldSelector_DuckdbExpression{DuckdbExpression: y.DuckDBExpression}
	}
	return res
}
