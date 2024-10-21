package fieldselectorpb

import (
	"fmt"
	"regexp"
	"slices"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// Resolve resolves a field selector against a list of all available fields.
func Resolve(fs *runtimev1.FieldSelector, all []string) ([]string, error) {
	if fs == nil {
		return nil, fmt.Errorf("empty field selector")
	}

	if fs.Selector == nil {
		if fs.Invert {
			return all, nil
		}
		return nil, fmt.Errorf("empty field selector")
	}

	switch fs.Selector.(type) {
	case *runtimev1.FieldSelector_All:
		if fs.Invert {
			return nil, nil
		}
		return all, nil
	case *runtimev1.FieldSelector_Fields:
		// Check that all fields in the selector are present in all
		selectorFields := fs.GetFields().Values
		allMap := make(map[string]struct{}, len(all))
		for _, f := range all {
			allMap[f] = struct{}{}
		}
		for _, f := range selectorFields {
			_, ok := allMap[f]
			if !ok {
				return nil, fmt.Errorf("invalid field selector: field %q not found", f)
			}
		}

		// Not inverted – return the selectorFields
		if !fs.Invert {
			return selectorFields, nil
		}

		// Inverted – return all fields except those in selectorFields
		if len(all) == len(selectorFields) {
			// Optimization for exclude all
			return nil, nil
		}
		res := make([]string, 0, len(all)-len(selectorFields))
		for _, f := range all {
			if !slices.Contains(selectorFields, f) {
				res = append(res, f)
			}
		}
		return res, nil
	case *runtimev1.FieldSelector_Regex:
		r, err := regexp.Compile(fs.GetRegex())
		if err != nil {
			return nil, fmt.Errorf("invalid field selector regex: %w", err)
		}

		if fs.Invert {
			res := make([]string, 0, len(all))
			for _, f := range all {
				if !r.MatchString(f) {
					res = append(res, f)
				}
			}
			return res, nil
		}

		res := make([]string, 0, len(all))
		for _, f := range all {
			if r.MatchString(f) {
				res = append(res, f)
			}
		}
		return res, nil
	case *runtimev1.FieldSelector_DuckdbExpression:
		selectorFields, err := resolveDuckDBExpression(fs.GetDuckdbExpression(), all)
		if err != nil {
			return nil, fmt.Errorf("error evaluating DuckDB field selector expression: %w", err)
		}

		if !fs.Invert {
			return selectorFields, nil
		}

		// Invert the selector fields
		if len(all) == len(selectorFields) {
			// Optimization for exclude all
			return nil, nil
		}
		res := make([]string, 0, len(all)-len(selectorFields))
		for _, f := range all {
			if !slices.Contains(selectorFields, f) {
				res = append(res, f)
			}
		}
		return res, nil
	default:
		return nil, fmt.Errorf("invalid field selector %T", fs.Selector)
	}
}
