package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// Label keys that a translation file can set. They correspond to fields on the translated resource's spec.
const (
	TranslationLabelDisplayName = "display_name"
	TranslationLabelDescription = "description"
)

// TranslationJSON is the raw structure of a translation file, keyed by the name of the resource to translate.
type TranslationJSON map[string]TranslationResourceJSON

// TranslationResourceJSON is the set of translations for a single resource.
// The dimensions and measures blocks exist for authoring convenience; they are merged into one map in the spec.
type TranslationResourceJSON struct {
	DisplayName string                           `json:"display_name"`
	Description string                           `json:"description"`
	Dimensions  map[string]TranslationLabelsJSON `json:"dimensions"`
	Measures    map[string]TranslationLabelsJSON `json:"measures"`
}

// TranslationLabelsJSON is the set of translations for a dimension or measure.
type TranslationLabelsJSON struct {
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

// parseTranslation parses a translation file and adds the resulting resource to p.Resources.
// Unlike other resources, translations are JSON files and are not parsed through a Node:
// their top-level keys are resource names, so they can't be decoded into commonYAML,
// and there is no place to put a "type" in the format anyway.
func (p *Parser) parseTranslation(ctx context.Context, path string) error {
	data, err := p.Repo.Get(ctx, path)
	if err != nil {
		if os.IsNotExist(err) {
			// This is a dirty parse where the file disappeared during parsing.
			// The clear-and-rebuild behavior means we can safely skip it.
			return nil
		}
		return err
	}
	if len(data) > maxFileSize {
		return fmt.Errorf("size %d bytes exceeds max size of %d bytes", len(data), maxFileSize)
	}

	// Decoding with DisallowUnknownFields so that a typo or a misplaced block is an error instead of a silent no-op.
	dec := json.NewDecoder(strings.NewReader(data))
	dec.DisallowUnknownFields()
	tmp := TranslationJSON{}
	err = dec.Decode(&tmp)
	if err != nil {
		return fmt.Errorf("failed to parse translation file: %w", err)
	}

	// Iterating over sorted names so that errors and refs are stable across reparses.
	names := slices.Sorted(maps.Keys(tmp))

	spec := &runtimev1.TranslationSpec{Resources: make(map[string]*runtimev1.TranslationSpec_ResourceTranslation, len(tmp))}
	refs := make([]ResourceName, 0, len(tmp))
	for _, name := range names {
		res := tmp[name]

		rt := &runtimev1.TranslationSpec_ResourceTranslation{
			BaseTranslation: newTranslationLabels(runtimev1.TranslationSpec_LABELS_TYPE_BASE, res.DisplayName, res.Description),
		}

		for _, dim := range slices.Sorted(maps.Keys(res.Dimensions)) {
			labels := newTranslationLabels(runtimev1.TranslationSpec_LABELS_TYPE_DIMENSION, res.Dimensions[dim].DisplayName, res.Dimensions[dim].Description)
			if labels == nil {
				continue
			}
			if rt.SubTranslations == nil {
				rt.SubTranslations = make(map[string]*runtimev1.TranslationSpec_Labels)
			}
			rt.SubTranslations[dim] = labels
		}

		for _, measure := range slices.Sorted(maps.Keys(res.Measures)) {
			// The sub translations are keyed by name only, so a name can't be both a dimension and a measure.
			if _, ok := res.Dimensions[measure]; ok {
				return fmt.Errorf("%q: %q is translated as both a dimension and a measure", name, measure)
			}
			labels := newTranslationLabels(runtimev1.TranslationSpec_LABELS_TYPE_MEASURE, res.Measures[measure].DisplayName, res.Measures[measure].Description)
			if labels == nil {
				continue
			}
			if rt.SubTranslations == nil {
				rt.SubTranslations = make(map[string]*runtimev1.TranslationSpec_Labels)
			}
			rt.SubTranslations[measure] = labels
		}

		spec.Resources[name] = rt

		// The kind is unspecified because the file only gives us a name. It's resolved by inferAmbiguousRefs.
		refs = append(refs, ResourceName{Name: name})
	}

	// The resource name is the locale, derived from the file name. There is no way to override it.
	locale := filepath.Base(pathStem(path))

	r, err := p.insertResource(ResourceKindTranslation, locale, []string{path}, nil, refs...)
	if err != nil {
		return err
	}
	r.TranslationSpec = spec

	return nil
}

// newTranslationLabels builds a Labels message of the given type.
// It returns nil if no labels were set, so that a blank value doesn't read as "clear this field" downstream.
func newTranslationLabels(typ runtimev1.TranslationSpec_LabelsType, displayName, description string) *runtimev1.TranslationSpec_Labels {
	labels := make(map[string]string, 2)
	if displayName != "" {
		labels[TranslationLabelDisplayName] = displayName
	}
	if description != "" {
		labels[TranslationLabelDescription] = description
	}
	if len(labels) == 0 {
		return nil
	}
	return &runtimev1.TranslationSpec_Labels{Type: typ, Labels: labels}
}
