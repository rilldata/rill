package jsonschemautil

import (
	"encoding/json"
	"maps"
	"regexp"
	"slices"
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
			sub, subDefs := propertySchema(s, defs, key, 0)
			if sub == nil {
				continue
			}
			if nv, ch := coerceValue(sub, subDefs, val); ch {
				v[key] = nv
				changed = true
			}
		}
	case []any:
		items, itemDefs := itemsSchema(s, defs, 0)
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
		// A $ref composes with its sibling keywords rather than replacing them.
		// Following the ref discards the siblings, so if the node itself permits strings,
		// coercion could rewrite a string the schema accepts as-is; fail open instead.
		if s.Type == "string" || slices.Contains(s.Types, "string") {
			return nil, defs
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
	// allOf is a conjunction: the value must satisfy every branch,
	// so the allowed types are the intersection of the branches' known type sets.
	// Branches without a type constraint don't restrict the conjunction.
	if len(s.AllOf) > 0 {
		var types map[string]bool
		known := false
		for _, b := range s.AllOf {
			branchTypes, ok := effectiveTypes(b, defs, depth+1)
			if !ok {
				continue
			}
			if !known {
				types, known = branchTypes, true
				continue
			}
			for t := range types {
				if !branchTypes[t] {
					delete(types, t)
				}
			}
		}
		if known {
			return types, true
		}
	}
	// anyOf/oneOf is a disjunction: the allowed types are the union of the branches' type sets,
	// which is only known if every branch's type set is known.
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
// depth guards against cyclic combinator branches (e.g. an allOf branch that refs back to its parent),
// whose recursion is driven purely by the schema and would otherwise overflow the stack.
func propertySchema(s *jsonschema.Schema, defs map[string]*jsonschema.Schema, key string, depth int) (*jsonschema.Schema, map[string]*jsonschema.Schema) {
	if depth > maxRefHops {
		return nil, nil
	}
	if sub, ok := s.Properties[key]; ok {
		return sub, defs
	}
	// additionalProperties does not govern keys matched by patternProperties; fail open for those keys.
	// On an invalid pattern, also fail open and leave the error to schema validation.
	for pattern := range s.PatternProperties {
		if matched, err := regexp.MatchString(pattern, key); err != nil || matched {
			return nil, nil
		}
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
		sub, subDefs := propertySchema(b, branchDefs, key, depth+1)
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
// depth guards against cyclic combinator branches; see propertySchema.
func itemsSchema(s *jsonschema.Schema, defs map[string]*jsonschema.Schema, depth int) (*jsonschema.Schema, map[string]*jsonschema.Schema) {
	if depth > maxRefHops {
		return nil, nil
	}
	// prefixItems changes which elements `items` governs (only those after the prefix);
	// fail open rather than coercing tuple elements with the wrong schema.
	if len(s.PrefixItems) > 0 {
		return nil, nil
	}
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
		sub, subDefs := itemsSchema(b, branchDefs, depth+1)
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
