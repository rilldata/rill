package parser

import (
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"google.golang.org/protobuf/types/known/structpb"
)

// TranslationsYAML is the raw YAML structure of a sub-property for defining a set of translations for a resource.
type TranslationsYAML map[string]TranslationYAML

func (t *TranslationsYAML) Proto() (*runtimev1.Translations, error) {
	translations := &runtimev1.Translations{Translations: make(map[string]*runtimev1.Translations_Translation, len(*t))}
	for k, v := range *t {
		translation, err := v.Proto()
		if err != nil {
			return nil, err
		}
		translations.Translations[k] = translation
	}
	return translations, nil
}

type TranslationYAML struct {
	DisplayName string `yaml:"display_name"`
	Description string `yaml:"description"`
	Dimensions  map[string]*struct {
		DisplayName string `yaml:"display_name"`
		Description string `yaml:"description"`
	} `yaml:"dimensions"`
	Measures map[string]*struct {
		DisplayName    string         `yaml:"display_name"`
		Description    string         `yaml:"description"`
		FormatD3Locale map[string]any `yaml:"format_d3_locale"`
	} `yaml:"measures"`
}

func (t *TranslationYAML) Proto() (*runtimev1.Translations_Translation, error) {
	translation := &runtimev1.Translations_Translation{
		DisplayName: t.DisplayName,
		Description: t.Description,
	}

	if t.Dimensions != nil {
		translation.Dimensions = make(map[string]*runtimev1.Translations_DimensionTranslation)
		for k, v := range t.Dimensions {
			if v == nil {
				continue
			}

			translation.Dimensions[k] = &runtimev1.Translations_DimensionTranslation{
				DisplayName: v.DisplayName,
				Description: v.Description,
			}
		}
	}

	if t.Measures != nil {
		translation.Measures = make(map[string]*runtimev1.Translations_MeasureTranslation)

		for k, v := range t.Measures {
			if v == nil {
				continue
			}

			var formatD3Locale *structpb.Struct
			if v.FormatD3Locale != nil {
				st, err := pbutil.ToStruct(v.FormatD3Locale, nil)
				if err != nil {
					return nil, err
				}
				formatD3Locale = st
			}

			translation.Measures[k] = &runtimev1.Translations_MeasureTranslation{
				DisplayName:    v.DisplayName,
				Description:    v.Description,
				FormatD3Locale: formatD3Locale,
			}
		}
	}

	return translation, nil
}
