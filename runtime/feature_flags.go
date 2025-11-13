package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/parser"
	"golang.org/x/exp/maps"
)

// FeatureFlags finds and resolves the feature flags for the given instance ID and claims.
// It's designed for use in the backend. Use runtime.ResolveFeatureFlags for resolving flags that will be exposed to the UI.
func (r *Runtime) FeatureFlags(ctx context.Context, instanceID string, claims *SecurityClaims) (map[string]bool, error) {
	inst, err := r.Instance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	featureFlags, err := ResolveFeatureFlags(inst, claims.UserAttributes, false)
	if err != nil {
		return nil, err
	}
	return featureFlags, nil
}

// defaultFeatureFlags contains the default feature flag rules.
var defaultFeatureFlags = map[string]string{
	// Controls whether the export functionality is visible
	"exports": "true",
	// Controls visibility of the source data viewer table in Rill Cloud for metrics views
	"cloud_data_viewer": "false",
	// Controls visibility of the global dimension search feature
	"dimension_search": "false",
	// TODO: more info
	"two_tiered_navigation": "false",
	// Controls visibility of the RillTime syntax range picker
	"rill_time": "true",
	// Controls visibility of the public URL sharing option in dashboards
	"hide_public_url": "{{.user.embed}}",
	// TODO: more info
	"export_header": "false",
	// Controls visibility of alert creation functionality
	"alerts": "true",
	// Controls visibility of report creation functionality
	"reports": "true",
	// Controls visibility of theme switching between light/dark modes
	"dark_mode": "true",
	// Controls visibility of project-level chat functionality
	"chat": "true",
	// Controls visibility of dashboard-level chat functionality
	"dashboard_chat": "false",
	// Controls whether charts are rendered in AI chats
	"chat_charts": "true",
	// Controls whether to show/hide deploy related actions.
	"deploy": "true",
}

// ResolveFeatureFlags resolves feature flags for the given instance and the provided user attributes.
// Set the camelCase flag to true when the flags will be returned to the UI since it currently expects the keys in that format.
// NOTE: Currently, feature flags are mainly honored in the UI. They are only enforced on the backend in a few places.
func ResolveFeatureFlags(inst *drivers.Instance, userAttributes map[string]any, camelCase bool) (map[string]bool, error) {
	if userAttributes == nil {
		userAttributes = make(map[string]any)
	}

	templateData := parser.TemplateData{
		Environment: inst.Environment,
		User:        userAttributes,
		Variables:   inst.ResolveVariables(false),
	}

	mergedFeatureFlags := maps.Clone(defaultFeatureFlags)
	maps.Copy(mergedFeatureFlags, inst.FeatureFlags)

	featureFlags := make(map[string]bool)
	for k, v := range mergedFeatureFlags {
		var bv bool
		switch v {
		case "true": // Just an optimization
			bv = true
		case "false": // Just an optimization
			bv = false
		default:
			rv, err := parser.ResolveTemplate(v, templateData, false)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve feature flag %q with template %q: %w", k, v, err)
			}
			rv = strings.TrimSpace(rv)

			if rv != "" && rv != "<no value>" {
				bv, err = parser.EvaluateBoolExpression(rv)
				if err != nil {
					return nil, fmt.Errorf("failed to evaluate feature flag %q with template %q: %w", k, v, err)
				}
			}
		}

		if camelCase {
			k = strcase.ToLowerCamel(k)
		}
		featureFlags[k] = bv
	}

	return featureFlags, nil
}
