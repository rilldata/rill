package templates

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
)

// ResolveEnvVarNameForKey determines the env var name for a secret property key,
// using an explicit env var name when provided, or falling back to DRIVER_KEY format.
// This variant does not require a *drivers.PropertySpec; used by schema-based templates.
func ResolveEnvVarNameForKey(driverName, key, explicitEnvVarName string, existingEnv map[string]bool) string {
	var base string
	if explicitEnvVarName != "" {
		base = explicitEnvVarName
	} else {
		base = strings.ToUpper(driverName) + "_" + strings.ToUpper(key)
	}

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
