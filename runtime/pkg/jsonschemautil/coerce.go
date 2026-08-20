package jsonschemautil

import (
	"encoding/json"
	"maps"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
)

// maxRefHops caps `$ref` chain resolution to guard against cyclic references.
const maxRefHops = 32

// CoerceStringifiedJSON walks value alongside schema and, wherever the schema
// unambiguously expects an object or array but the value is a string containing valid JSON of that kind,
// replaces the string with the parsed value.
// It returns the (possibly mutated) value and whether anything changed.
// It never errors: on any ambiguity or parse failure it leaves the value untouched
// so that downstream schema validation produces its normal error.
// It exists to work around MCP clients that JSON-encode object-typed tool arguments as strings
// (see https://github.com/anthropics/claude-code/issues/25865).
func CoerceStringifiedJSON(schema *jsonschema.Schema, value any) (any, bool) {
	return coerceValue(schema, nil, value)
}

// coerceValue implements CoerceStringifiedJSON for a single value and its subschema.
// defs holds the `$defs` visible at this point in the schema tree;
// resolveSchema extends it with the current schema's own `$defs`,
// which is required because schemas built with jsonschema.ForOptions.TypeSchemas
// carry `$defs` nested inside property subschemas rather than at the root.
func coerceValue(s *jsonschema.Schema, defs map[string]*jsonschema.Schema, value any) (any, bool) {
	s, defs = resolveSchema(s, defs)
	if s == nil || value == nil {
		return value, false
	}

	changed := false

	// If the value is a string but the schema only allows objects (or arrays), attempt to JSON-decode it.
	// The type check is deliberately conservative: if the schema also allows strings (e.g. anyOf [object, string])
	// or has no type constraint at all, the value is left untouched.
	if str, ok := value.(string); ok {
		types, known := effectiveTypes(s, defs, 0)
		if known {
			var want string
			if wantsOnly(types, "object") {
				want = "{"
			} else if wantsOnly(types, "array") {
				want = "["
			}
			if want != "" {
				if parsed, ok := decodeJSONString(str, want); ok {
					value = parsed
					changed = true
				}
			}
		}
	}

	// Recurse into containers. Mutating in place is safe because the caller owns the freshly decoded value.
	switch v := value.(type) {
	case map[string]any:
		for key, val := range v {
			sub, subDefs := propertySchema(s, defs, key)
			if sub == nil {
				continue
			}
			if nv, ch := coerceValue(sub, subDefs, val); ch {
				v[key] = nv
				changed = true
			}
		}
	case []any:
		items, itemDefs := itemsSchema(s, defs)
		if items != nil {
			for i, item := range v {
				if nv, ch := coerceValue(items, itemDefs, item); ch {
					v[i] = nv
					changed = true
				}
			}
		}
	}

	return value, changed
}

// resolveSchema merges the schema's `$defs` into the visible scope and follows local `$ref` chains.
// It returns a nil schema if a ref cannot be resolved or points outside `#/$defs/`.
func resolveSchema(s *jsonschema.Schema, defs map[string]*jsonschema.Schema) (*jsonschema.Schema, map[string]*jsonschema.Schema) {
	for range maxRefHops {
		if s == nil {
			return nil, defs
		}
		if len(s.Defs) > 0 {
			merged := make(map[string]*jsonschema.Schema, len(defs)+len(s.Defs))
			maps.Copy(merged, defs)
			maps.Copy(merged, s.Defs)
			defs = merged
		}
		if s.Ref == "" {
			return s, defs
		}
		name, ok := strings.CutPrefix(s.Ref, "#/$defs/")
		if !ok {
			return nil, defs
		}
		s = defs[name]
	}
	return nil, defs
}

// effectiveTypes returns the set of JSON types the schema allows,
// and whether that set could be determined.
// A schema without any type constraint (such as a free-form value field) returns known=false,
// which blocks coercion.
func effectiveTypes(s *jsonschema.Schema, defs map[string]*jsonschema.Schema, depth int) (map[string]bool, bool) {
	if depth > maxRefHops {
		return nil, false
	}
	s, defs = resolveSchema(s, defs)
	if s == nil {
		return nil, false
	}
	if s.Type != "" {
		return map[string]bool{s.Type: true}, true
	}
	if len(s.Types) > 0 {
		types := make(map[string]bool, len(s.Types))
		for _, t := range s.Types {
			types[t] = true
		}
		return types, true
	}
	if len(s.AnyOf) > 0 || len(s.OneOf) > 0 {
		types := make(map[string]bool)
		for _, branches := range [][]*jsonschema.Schema{s.AnyOf, s.OneOf} {
			for _, b := range branches {
				branchTypes, known := effectiveTypes(b, defs, depth+1)
				if !known {
					return nil, false
				}
				maps.Copy(types, branchTypes)
			}
		}
		return types, true
	}
	return nil, false
}

// wantsOnly reports whether the type set allows kind and nothing else except null.
func wantsOnly(types map[string]bool, kind string) bool {
	if !types[kind] {
		return false
	}
	for t := range types {
		if t != kind && t != "null" {
			return false
		}
	}
	return true
}

// decodeJSONString parses a string as a single JSON value starting with the given delimiter ("{" or "[").
// It uses json.Number to preserve integer fidelity across a later re-marshal.
func decodeJSONString(s, delim string) (any, bool) {
	trimmed := strings.TrimSpace(s)
	if !strings.HasPrefix(trimmed, delim) {
		return nil, false
	}
	dec := json.NewDecoder(strings.NewReader(trimmed))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, false
	}
	// Reject trailing content after the decoded value (dec.More() misses trailing "}" or "]").
	if strings.TrimSpace(trimmed[dec.InputOffset():]) != "" {
		return nil, false
	}
	return v, true
}

// propertySchema returns the subschema for a map key, along with the `$defs` scope it should be evaluated in,
// or nil if the schema does not unambiguously define one.
func propertySchema(s *jsonschema.Schema, defs map[string]*jsonschema.Schema, key string) (*jsonschema.Schema, map[string]*jsonschema.Schema) {
	if sub, ok := s.Properties[key]; ok {
		return sub, defs
	}
	if s.AdditionalProperties != nil {
		return s.AdditionalProperties, defs
	}
	// Search allOf branches; only trust the result if exactly one branch defines the key.
	// allOf is a conjunction, so a property schema found in one branch is binding.
	// anyOf/oneOf are deliberately not searched: a disjunctive branch that omits the key
	// still accepts it by default, so a definition found in one branch is not unambiguous.
	var found *jsonschema.Schema
	var foundDefs map[string]*jsonschema.Schema
	for _, b := range s.AllOf {
		b, branchDefs := resolveSchema(b, defs)
		if b == nil {
			continue
		}
		sub, subDefs := propertySchema(b, branchDefs, key)
		if sub == nil {
			continue
		}
		if found != nil {
			return nil, nil
		}
		found, foundDefs = sub, subDefs
	}
	return found, foundDefs
}

// itemsSchema returns the subschema for array elements, along with the `$defs` scope it should be evaluated in,
// or nil if the schema does not unambiguously define one.
func itemsSchema(s *jsonschema.Schema, defs map[string]*jsonschema.Schema) (*jsonschema.Schema, map[string]*jsonschema.Schema) {
	if s.Items != nil {
		return s.Items, defs
	}
	// Search allOf branches only; see propertySchema for why anyOf/oneOf are excluded.
	var found *jsonschema.Schema
	var foundDefs map[string]*jsonschema.Schema
	for _, b := range s.AllOf {
		b, branchDefs := resolveSchema(b, defs)
		if b == nil {
			continue
		}
		sub, subDefs := itemsSchema(b, branchDefs)
		if sub == nil {
			continue
		}
		if found != nil {
			return nil, nil
		}
		found, foundDefs = sub, subDefs
	}
	return found, foundDefs
}
