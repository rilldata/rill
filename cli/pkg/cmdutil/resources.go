package cmdutil

import (
	"fmt"
	"strings"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/rilldata/rill/runtime"
)

// ParseResourceStrings converts user supplied resource strings into proto resource names while deduping and validating input.
func ParseResourceStrings(explores, canvases []string) ([]*adminv1.ResourceName, error) {
	if len(explores) == 0 && len(canvases) == 0 {
		return nil, nil
	}

	res := make([]*adminv1.ResourceName, 0, len(explores)+len(canvases))
	exploreMap := make(map[string]struct{}, len(explores))
	canvasMap := make(map[string]struct{}, len(canvases))

	for _, raw := range explores {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}

		if _, ok := exploreMap[name]; ok {
			continue
		}
		exploreMap[name] = struct{}{}

		res = append(res, &adminv1.ResourceName{
			Type: runtime.ResourceKindExplore,
			Name: name,
		})
	}

	for _, raw := range canvases {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}

		if _, ok := canvasMap[name]; ok {
			continue
		}
		canvasMap[name] = struct{}{}

		res = append(res, &adminv1.ResourceName{
			Type: runtime.ResourceKindCanvas,
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
