package ai

import (
	"context"

	"github.com/rilldata/rill/runtime"
)

// PreloadSkills exposes preloadSkills to the external test package.
var PreloadSkills = preloadSkills

// PersonalCanvasPath exposes personalCanvasPath to the external test package.
var PersonalCanvasPath = personalCanvasPath

// PersonalCanvasStatus exposes the status write_personal_canvas reports for a personal canvas to the external test package.
func PersonalCanvasStatus(ctx context.Context, rt *runtime.Runtime, path string) ([]map[string]any, string, error) {
	resources, parseErr, _, err := (&WriteFile{Runtime: rt}).reconcileAndGetStatus(ctx, path, true)
	return resources, parseErr, err
}
