package cmdutil

import (
	"fmt"
	"strings"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime"
)

// ParseResourceStrings converts user supplied resource strings (formatted as kind/name)
// into proto resource names while deduping and validating input.
func ParseResourceStrings(inputs []string) ([]*adminv1.ResourceName, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("at least one resource must be specified")
	}

	res := make([]*adminv1.ResourceName, 0, len(inputs))
	seen := make(map[string]struct{}, len(inputs))

	for _, raw := range inputs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		var kind, name string
		if strings.Contains(raw, "/") {
			parts := strings.SplitN(raw, "/", 2)
			kind = strings.TrimSpace(parts[0])
			if len(parts) > 1 {
				name = strings.TrimSpace(parts[1])
			}
		} else {
			return nil, fmt.Errorf("invalid resource %q, expected format kind/name", raw)
		}

		if kind == "" || name == "" {
			return nil, fmt.Errorf("invalid resource %q, expected format kind/name", raw)
		}

		kind = runtime.ResourceKindFromShorthand(kind)
		if !runtime.IsKnownResourceKind(kind) {
			return nil, fmt.Errorf("unknown resource kind %q in resource %q", kind, raw)
		}

		key := kind + "|" + strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		res = append(res, &adminv1.ResourceName{
			Type: kind,
			Name: name,
		})
	}

	if len(res) == 0 {
		return nil, fmt.Errorf("no valid resources were provided")
	}

	return res, nil
}

// FormatResourceNames renders a human-readable list of resources.
func FormatResourceNames(resources []*adminv1.ResourceName) string {
	if len(resources) == 0 {
		return ""
	}
	parts := make([]string, len(resources))
	for i, r := range resources {
		parts[i] = fmt.Sprintf("%s/%s", r.Type, r.Name)
	}
	return strings.Join(parts, ", ")
}
