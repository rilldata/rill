package templates

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

// ResolveEnvVarName determines the env var name for a secret property, resolving conflicts
// by appending _1, _2, etc. when the base name already exists in existingEnv.
func ResolveEnvVarName(driverName string, propSpec *drivers.PropertySpec, existingEnv map[string]bool) string {
	var base string
	if propSpec.EnvVarName != "" {
		base = propSpec.EnvVarName
	} else {
		// Fallback: DRIVER_KEY format (SCREAMING_SNAKE_CASE)
		base = strings.ToUpper(driverName) + "_" + strings.ToUpper(propSpec.Key)
	}

	// Check for conflicts
	candidate := base
	for i := 1; existingEnv[candidate]; i++ {
		candidate = fmt.Sprintf("%s_%d", base, i)
	}
	return candidate
}

// ReadEnvKeys parses a .env file into a set of key names.
func ReadEnvKeys(ctx context.Context, repo drivers.RepoStore) map[string]bool {
	keys := make(map[string]bool)
	content, err := repo.Get(ctx, ".env")
	if err != nil {
		return keys
	}
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.IndexByte(line, '='); idx > 0 {
			keys[line[:idx]] = true
		}
	}
	return keys
}
